/*
 * SPDX-FileCopyrightText: Copyright 2019, 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"errors"
)

type DeviceId uint8
type Command uint8

const (
	CmdGridOn           = Command(0x02)
	CmdGridOff          = Command(0x03)
	CmdSetPowerStandard = Command(0x05)
	CmdPing             = Command(0x06)
	CmdGetInformation   = Command(0xa1)
	CmdGetPowerCurve    = Command(0xa3)
	CmdSelectPowerCurve = Command(0xa4)
	CmdUpdatePowerCurve = Command(0xaa)
	CmdLog              = Command(0xc1)
)

// Checksum mismatch in frame
var ChecksumError = errors.New("Checksum error")

// Illegal frame received
var IllegalFrameError = errors.New("Illegal frame")

type Frame struct {
	Device  DeviceId
	Command Command
	Length  uint8
	Data    []byte
}

type BusInterface interface {
	ReadFrame() (*Frame, error)
	ReadAckFrame() (*Frame, error)
	WriteFrame(frame *Frame) error
	WriteAck(dev DeviceId, cmd Command) error
}
