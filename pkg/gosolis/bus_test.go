/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"bytes"
	"io"
	"testing"
)

func testChecksum(t *testing.T, value []byte, expected uint8) {
	if sum := calcChecksum(value); sum != expected {
		t.Errorf("calcChecksum(%v) = %d; want %d.", value, sum, expected)
	}
}

func TestChecksum(t *testing.T) {
	testChecksum(t, []byte{}, 0)
	testChecksum(t, []byte{1, 2}, 3)
	testChecksum(t, []byte{127, 1, 1}, 129)
	testChecksum(t, []byte{255, 255}, 254)
}

func testWaitForStart(t *testing.T, value []byte, expected error) *bytes.Buffer {
	buf := bytes.NewBuffer(value)
	s := NewBus(buf)
	e := s.waitForStart()
	if e != expected {
		t.Errorf("Error in waitForStart(%v) = %v; want %v",
			value, e, expected)
	}
	return buf
}

func TestWaitForStart(t *testing.T) {
	testWaitForStart(t, []byte{}, io.EOF)
	testWaitForStart(t, []byte{0, 1, 2}, io.EOF)

	buf := testWaitForStart(t, []byte{0, 0x7e, 2}, nil)
	b := buf.Bytes()
	if len(b) != 1 || b[0] != 2 {
		t.Errorf("Unexpected buffer content %v; want [ 2 ]", b)
	}
}

func readFrameBytes(value []byte) (*Frame, error) {
	buf := bytes.NewBuffer(value)
	s := NewBus(buf)
	return s.ReadFrame()
}

func compareFrames(t *testing.T, f *Frame, expected *Frame) (match bool) {
	if f == nil && expected == nil {
		return true
	}

	if f != nil && expected == nil {
		t.Errorf("Got a frame, but didn't expect one")
		return false
	} else if f == nil && expected != nil {
		t.Errorf("Expected a frame, but didn't get one")
		return false
	}

	match = true
	if f.Device != expected.Device {
		t.Errorf("Unexpected device %d; want %d.",
			f.Device, expected.Device)
		match = false
	}

	if f.Command != expected.Command {
		t.Errorf("Unexpected command %d; want %d.",
			f.Command, expected.Command)
		match = false
	}

	if f.Length != expected.Length {
		t.Errorf("Unexpected length %d; want %d.",
			f.Length, expected.Length)
		match = false
	}

	if f.Data != nil && expected.Data != nil {
		if bytes.Compare(f.Data, expected.Data) != 0 {
			t.Error("ReadFrame data mismatch")
			match = false
		}
	} else if f.Data != nil || expected.Data != nil {
		t.Error("ReadFrame data mismatch: ",
			f.Data, expected.Data)
		match = false
	}

	return
}

func TestReadFrame(t *testing.T) {
	t.Log("Testing ReadFrame(CmdGridOff)")
	grid_off_frame := make([]byte, 55)
	copy(grid_off_frame, []byte{0x7e, 0x01, 0x03, 0x00})
	grid_off_frame[54] = byte(0x01 + 0x03)
	if f, e := readFrameBytes(grid_off_frame); e == nil {
		expected := Frame{DeviceId(0x01), CmdGridOff, 0, make([]byte, 50)}
		if !compareFrames(t, f, &expected) {
			t.Error("Frame mismatch")
		}
	} else {
		t.Error("ReadFrame failed: ", e)
	}

	// Fail waiting for start byte
	if f, e := readFrameBytes(grid_off_frame[1:]); f != nil || e != io.EOF {
		t.Errorf("ReadFrame returned %v, %v; want nil, EOF", f, e)
	}

	// Short read when reading packet data / checksum
	if f, e := readFrameBytes(grid_off_frame[0:54]); f != nil || e == nil {
		t.Errorf("ReadFrame returned %v, %v; want nil, EOF", f, e)
	}

	// Illegal checksum
	grid_off_frame[54] = 0
	if f, e := readFrameBytes(grid_off_frame); f == nil || e != ChecksumError {
		t.Errorf("ReadFrame returned %v, %v; want !nil, ChecksumError", f, e)
	}
}

func TestReadAckFrame(t *testing.T) {
	buf := bytes.Buffer{}
	s := NewBus(&buf)

	// Valid ACK
	buf.Write([]byte{0x7e, 0x01, 0x02, 0x00})
	if f, e := s.ReadAckFrame(); f != nil && e == nil {
		expected := Frame{0x01, 0x02, 0x00, nil}
		if !compareFrames(t, f, &expected) {
			t.Errorf("Frame mismatch")
		}
		if buf.Len() != 0 {
			t.Error("Data left in buffer")
		}
	} else {
		t.Errorf("ReadAckFrame returned %v, %v; want !nil, nil.", f, e)
	}

	// No start marker
	buf.Write([]byte{0x00})
	if f, e := s.ReadAckFrame(); f != nil || e != io.EOF {
		t.Errorf("ReadAckFrame returned %v, %v; want nil, EOF", f, e)
	}
	if buf.Len() != 0 {
		t.Error("Data left in buffer")
	}

	// Illegal ack frame (len != 0)
	buf.Write([]byte{0x7e, 0x01, 0x02, 0xff})
	if f, e := s.ReadAckFrame(); f == nil || e != IllegalFrameError {
		t.Errorf("ReadAckFrame returned %v, %v; want !nil, IllegalFrameError", f, e)
	}
	if buf.Len() != 0 {
		t.Error("Data left in buffer")
	}

	// Short read when reading ack frame
	buf.Write([]byte{0x7e, 0x01, 0x02})
	if f, e := s.ReadAckFrame(); f != nil || e == nil {
		t.Errorf("ReadAckFrame returned %v, %v; want !nil, EOF", f, e)
	}
	if buf.Len() != 0 {
		t.Error("Data left in buffer")
	}
}

func TestWriteFrame(t *testing.T) {
	buf := bytes.Buffer{}
	s := NewBus(&buf)

	f := Frame{0x01, 0x42, 2, []byte{3, 4, 5}}
	if e := s.WriteFrame(&f); e == nil {
		b := buf.Bytes()

		expected := make([]byte, 55)
		copy(expected, []byte{0x7e, 0x01, 0x42, 2, 3, 4, 5})
		expected[54] = 0x01 + 0x42 + 2 + 3 + 4 + 5

		if bytes.Compare(b, expected) != 0 {
			t.Errorf("WriteFrame emitted %#v; expected %#v",
				b, expected)
		}
	} else {
		t.Errorf("WriteFrame returned %v; want nil", e)
	}

	buf.Reset()
	f = Frame{0x01, 0x42, 2, make([]byte, 51)}
	if e := s.WriteFrame(&f); e != IllegalFrameError {
		t.Errorf("WriteFrame returned %v; want IllegalFrameError", e)
	}
}

func TestWriteAck(t *testing.T) {
	buf := bytes.Buffer{}
	s := NewBus(&buf)
	if e := s.WriteAck(DeviceId(0x10), Command(0x11)); e != nil {
		t.Error(e)
	}

	b := buf.Bytes()
	if len(b) != 4 ||
		b[0] != 0x7e || b[1] != 0x10 ||
		b[2] != 0x11 || b[3] != 0x00 {
		t.Errorf("Unexpected buffer content %v; want "+
			"[ 0x7e, 0x10, 0x11, 0x00]", buf)
	}
}
