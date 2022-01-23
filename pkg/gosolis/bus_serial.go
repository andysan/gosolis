/*
 * SPDX-FileCopyrightText: Copyright 2019, 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"io"
)

const startByte = byte(0x7e)

// Maximum data length in a frame
const maxDataLength = 50

// Frame length excluding start byte
const frameLength = 54

// Length of an acknowledgement frame excluding start byte
const ackFrameLength = 3

type SerialBus struct {
	port io.ReadWriter
}

// Ensure that we satisfy the BusInterface interface
var _ BusInterface = &SerialBus{}

func calcChecksum(pkt []byte) uint8 {
	checksum := uint8(0)
	for _, c := range pkt {
		checksum += uint8(c)
	}

	return checksum
}

func (b *SerialBus) waitForStart() error {
	buf := make([]byte, 1)
	for {
		if _, err := io.ReadFull(b.port, buf); err != nil {
			return err
		}

		if buf[0] == startByte {
			return nil
		}
	}
}

// Instantiate a new bus interface using a ReadWriter interface connected to a
// RS485 port.
func NewSerialBus(port io.ReadWriter) *SerialBus {
	return &SerialBus{port}
}

// Read a data fram from the inverter and return a frame. This
// function may fail with ChecksumError and still return a frame.
func (b *SerialBus) ReadFrame() (*Frame, error) {
	if e := b.waitForStart(); e != nil {
		return nil, e
	}

	buf := make([]byte, frameLength)
	// All frames contain at least a device ID, command, and
	// length byte
	if _, err := io.ReadFull(b.port, buf); err != nil {
		return nil, err
	}

	frame := Frame{
		DeviceId(buf[0]), Command(buf[1]), buf[2],
		make([]byte, maxDataLength)}
	copy(frame.Data, buf[3:maxDataLength])

	// Got the entire frame, verify checksum
	if calcChecksum(buf[:len(buf)-1]) != buf[len(buf)-1] {
		return &frame, ChecksumError
	}

	return &frame, nil
}

func (b *SerialBus) ReadAckFrame() (*Frame, error) {
	if e := b.waitForStart(); e != nil {
		return nil, e
	}

	buf := make([]byte, ackFrameLength)
	if _, err := io.ReadFull(b.port, buf); err != nil {
		return nil, err
	}

	frame := Frame{DeviceId(buf[0]), Command(buf[1]), buf[2], nil}
	if frame.Length != 0 {
		return &frame, IllegalFrameError
	} else {
		return &frame, nil
	}
}

func (b *SerialBus) WriteFrame(frame *Frame) error {
	return b.writeFrame(frame.Device, frame.Command, frame.Length, frame.Data)
}

func (b *SerialBus) writeFrame(dev DeviceId, cmd Command, length uint8, data []byte) error {
	// Data won't fit in frame
	if len(data) > maxDataLength {
		return IllegalFrameError
	}

	// Reserve an additional byte for the start marker
	buf := make([]byte, frameLength+1)

	copy(buf, []byte{startByte, byte(dev), byte(cmd), byte(length)})
	copy(buf[4:], data)

	buf[len(buf)-1] = calcChecksum(buf[1 : len(buf)-1])
	_, err := b.port.Write(buf)
	return err
}

func (b *SerialBus) WriteAck(dev DeviceId, cmd Command) error {
	buf := []byte{startByte, byte(dev), byte(cmd), byte(0)}

	_, err := b.port.Write(buf)
	return err
}
