package unp

import (
	"bytes"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestDeconstructSize1(c *C) {
	var buffer bytes.Buffer
	unpi := New(1, &buffer)
	//empty payload
	frame := &Frame{C_SRSP, S_SAPI, 0, []byte{}}
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x00, 0x66, 0x00, 0x66})

	//nil payload
	frame = &Frame{C_SRSP, S_SAPI, 0, nil}
	buffer.Reset()
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x00, 0x66, 0x00, 0x66})

	//non empty payload
	frame = &Frame{C_SREQ, S_SAPI, 0, []byte{0x00, 0x01, 0x02}}
	buffer.Reset()
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x03, 0x26, 0x00, 0x00, 0x01, 0x02, 0x26})
}

func (s *MySuite) TestConstructSize1(c *C) {
	var buffer bytes.Buffer
	unpi := New(1, &buffer)

	//empty payload
	buffer.Write([]byte{0xfe, 0x00, 0x66, 0x00, 0x66})
	frame, _ := unpi.ReadFrame()
	c.Assert(frame, DeepEquals, &Frame{C_SRSP, S_SAPI, 0, []byte{}})

	//non empty payload
	buffer.Reset()
	buffer.Write([]byte{0xfe, 0x03, 0x26, 0x00, 0x00, 0x01, 0x02, 0x26})
	frame, _ = unpi.ReadFrame()
	c.Assert(frame, DeepEquals, &Frame{C_SREQ, S_SAPI, 0, []byte{0x00, 0x01, 0x02}})
}

func (s *MySuite) TestDeconstructSize2(c *C) {
	var buffer bytes.Buffer
	unpi := New(2, &buffer)

	//empty payload
	frame := &Frame{C_SRSP, S_SAPI, 0, []byte{}}
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x00, 0x00, 0x66, 0x00, 0x66})

	//nil payload
	frame = &Frame{C_SRSP, S_SAPI, 0, nil}
	buffer.Reset()
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x00, 0x00, 0x66, 0x00, 0x66})

	//non empty payload
	frame = &Frame{C_SREQ, S_SAPI, 0, []byte{0x00, 0x01, 0x02}}
	buffer.Reset()
	unpi.WriteFrame(frame)
	c.Assert(buffer.Bytes(), DeepEquals, []byte{0xfe, 0x00, 0x03, 0x26, 0x00, 0x00, 0x01, 0x02, 0x26})
}

func (s *MySuite) TestConstructSize2(c *C) {
	var buffer bytes.Buffer
	unpi := New(2, &buffer)

	//empty payload
	buffer.Write([]byte{0xfe, 0x00, 0x00, 0x66, 0x00, 0x66})
	frame, _ := unpi.ReadFrame()
	c.Assert(frame, DeepEquals, &Frame{C_SRSP, S_SAPI, 0, []byte{}})

	//non empty payload
	buffer.Reset()
	buffer.Write([]byte{0xfe, 0x00, 0x03, 0x26, 0x00, 0x00, 0x01, 0x02, 0x26})
	frame, _ = unpi.ReadFrame()
	c.Assert(frame, DeepEquals, &Frame{C_SREQ, S_SAPI, 0, []byte{0x00, 0x01, 0x02}})
}
