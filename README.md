# Unified Network Processor (UNP) Interface

[![Build Status](https://cloud.drone.io/api/badges/dyrkin/unp-go/status.svg?ref=/refs/heads/master)](https://cloud.drone.io/dyrkin/unp-go)

## Overview

This is Go implementation of TI's Unified Network Processor Interface.
It is
> used for establishing a serial data link between a TI SoC and external MCUs or PCs. This is mainly used by TI's network processor solutions.

I tested it with cc253**1**, but it might work with cc253**X**  
More info about UNPI can be located [here](http://processors.wiki.ti.com/index.php/Unified_Network_Processor_Interface)

To use it you need to provide a reference to a serial port:

```go
import (
	"go.bug.st/serial.v1"
	"github.com/dyrkin/unp-go"
)

func main() {
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open("/dev/tty.usbmodem14101", mode)
	if err != nil {
		log.Fatal(err)
	}
	port.SetRTS(true)

	u := unp.New(1, port)
}
```

Than you can able to read an write unp frames from it:

```go
//read from serial
frame, err := u.ReadFrame()

........

//write to serial		
frame = &Frame{CommandType:C_SREQ, Subsystem:S_SAPI, Command:0, Payload:[]byte{0x00, 0x01, 0x02}}
u.WriteFrame(frame)
```