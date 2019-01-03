package unpi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	transceiver io.ReadWriter
}

type Frame struct {
	commandType CommandType
	subsystem   Subsystem
	command     byte
	payload     []byte
}

const SOF byte = 0xFE

func New(size uint8, transmitter io.ReadWriter) *Unpi {
	return &Unpi{size, transmitter}
}

func (u *Unpi) Write(frame *Frame) (err error) {
	var writer bytes.Buffer

	cmd0 := ((byte(frame.commandType << 5)) & 0xE0) | (byte(frame.subsystem) & 0x1F)
	cmd1 := frame.command
	if err = writer.WriteByte(SOF); err != nil {
		return
	}
	if u.size == 1 {
		len := byte(len(frame.payload))
		if err = writer.WriteByte(len); err != nil {
			return
		}
	} else {
		length := len(frame.payload)
		if err = writer.WriteByte(byte(length >> 8)); err != nil {
			return
		}
		if err = writer.WriteByte(byte(length & 0xff)); err != nil {
			return
		}
	}
	if err = writer.WriteByte(cmd0); err != nil {
		return
	}
	if err = writer.WriteByte(cmd1); err != nil {
		return
	}
	if _, err = writer.Write(frame.payload); err != nil {
		return
	}
	fcs := checksum(writer.Bytes()[1:])
	if err = writer.WriteByte(fcs); err != nil {
		return
	}
	u.transceiver.Write(writer.Bytes())
	return
}

func (u *Unpi) Read() (frame *Frame, err error) {
	var b byte
	var buf [1]byte
	var checksumBuffer bytes.Buffer

	var read = func() {
		_, err = io.ReadFull(u.transceiver, buf[:])
		b = buf[0]
		checksumBuffer.WriteByte(b)
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
