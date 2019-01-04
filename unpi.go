package unpi

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/dyrkin/composer"
)

type CommandType byte

const (
	C_POLL CommandType = iota
	C_SREQ
	C_AREQ
	C_SRSP
	C_RES0
	C_RES1
	C_RES2
	C_RES3
)

type Subsystem byte

const (
	S_RES0 Subsystem = iota
	S_SYS
	S_MAC
	S_NWK
	S_AF
	S_ZDO
	S_SAPI
	S_UTIL
	S_DBG
	S_APP
	S_RCAF
	S_RCN
	S_RCN_CLIENT
	S_BOOT
	S_ZIPTEST
	S_DEBUG
	S_PERIPHERALS
	S_NFC
	S_PB_NWK_MGR
	S_PB_GW
	S_PB_OTA_MGR
	S_BLE_SPNP
	S_BLE_HCI
	S_RESV01
	S_RESV02
	S_RESV03
	S_RESV04
	S_RESV05
	S_RESV06
	S_RESV07
	S_RESV08
	S_SRV_CTR
)

type Unpi struct {
	size        uint8
	Transceiver io.ReadWriter
	incoming    chan byte
	errors      chan error
}

type Frame struct {
	CommandType CommandType
	Subsystem   Subsystem
	Command     byte
	Payload     []byte
}

const SOF byte = 0xFE

func New(size uint8, transmitter io.ReadWriter) *Unpi {
	u := &Unpi{size, transmitter, make(chan byte), make(chan error)}
	go u.byteReader()
	return u
}

func (u *Unpi) Write(frame *Frame) error {
	rendered := u.Render(frame)
	_, err := u.Transceiver.Write(rendered)
	return err
}

func (u *Unpi) Render(frame *Frame) []byte {
	cmp := composer.New()
	cmd0 := ((byte(frame.CommandType << 5)) & 0xE0) | (byte(frame.Subsystem) & 0x1F)
	cmd1 := frame.Command
	cmp.Byte(SOF)
	len := len(frame.Payload)
	if u.size == 1 {
		cmp.Uint8(uint8(len))
	} else {
		cmp.Uint16be(uint16(len))
	}
	cmp.Byte(cmd0).Byte(cmd1).Bytes(frame.Payload)
	fcs := checksum(cmp.Make()[1:])
	cmp.Byte(fcs)
	return cmp.Make()
}

func (u *Unpi) byteReader() {
	var buf [1]byte
	for {
		n, err := io.ReadFull(u.Transceiver, buf[:])
		if n > 0 {
			u.incoming <- buf[0]
		} else if err != io.EOF {
			u.errors <- err
		}
	}

}

func (u *Unpi) Read() (frame *Frame, err error) {
	var b byte
	var checksumBuffer bytes.Buffer

	var read = func() {
		select {
		case b = <-u.incoming:
			checksumBuffer.WriteByte(b)
		case err = <-u.errors:
		}
	}
	if read(); err != nil {
		return
	}
	if b != SOF {
		return nil, errors.New("Invalid start of frame")
	}
	if read(); err != nil {
		return
	}

	var payloadLength uint16
	if u.size == 1 {
		payloadLength = uint16(b)
	} else {
		b1 := uint16(b) << 8
		if read(); err != nil {
			return
		}
		payloadLength = b1 | uint16(b)
	}
	if read(); err != nil {
		return
	}
	cmd0 := b
	if read(); err != nil {
		return
	}
	cmd1 := b
	payload := make([]byte, payloadLength)
	for i := 0; i < int(payloadLength); i++ {
		if read(); err != nil {
			return
		}
		payload[i] = b
	}
	checksumBytes := checksumBuffer.Bytes()
	if read(); err != nil {
		return
	}
	fcs := b
	csum := checksum(checksumBytes[1:])
	if fcs != csum {
		err = fmt.Errorf("Invalid checksum. Expected: %b, actual: %b", fcs, csum)
		return
	}
	commandType := (cmd0 & 0xE0) >> 5
	subsystem := cmd0 & 0x1F
	return &Frame{CommandType(commandType), Subsystem(subsystem), cmd1, payload}, nil
}

func checksum(buf []byte) byte {
	fcs := byte(0)
	for i := 0; i < len(buf); i++ {
		fcs ^= buf[i]
	}
	return fcs
}
