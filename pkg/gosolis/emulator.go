/*
 * SPDX-FileCopyrightText: Copyright 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"fmt"
)

type DeviceEmulator struct {
	bus BusInterface
	dev DeviceId

	DeviceInformation DeviceInformation
}

var defaultEmulatedDeviceInformation DeviceInformation = DeviceInformation{
	Inputs: []InputStatus{
		InputStatus{Voltage: 42, Current: 1},
	},
	Grid: GridInformation{
		Voltage:       240.0,
		Current:       0.4,
		Frequency:     50.0,
		PowerStandard: PowerStandardG59G83,
		GridStatus:    GridStatus(0xbe),
	},
	Production: ProductionInformation{
		Total:     123456,
		Month:     256,
		LastMonth: 512,
		Today:     42.0,
		Yesterday: 31.0,
	},
	Temperature: 21.1,
	Product:     DeviceProduct(0x96),
	SWVersion:   DeviceVersion(0xf),
	SerialNo:    DeviceSerialNumber{1, 2, 3, 4, 5, 6, 7, 8},
	Status:      DeviceStatus(0x0001),
	Error:       DeviceError(0x0000),
	PowerCurve:  PowerCurve(0x04),
}

func NewDeviceEmulator(bus BusInterface, dev DeviceId) *DeviceEmulator {
	return &DeviceEmulator{bus, dev, defaultEmulatedDeviceInformation}
}

func (d *DeviceEmulator) Run() {
	commandDispatchers := map[Command]func(frame *Frame) error{
		CmdGridOn:           d.cmdAckIgnored,
		CmdGridOff:          d.cmdAckIgnored,
		CmdSetPowerStandard: d.cmdAckIgnored,
		CmdPing:             d.cmdAckIgnored,
		CmdGetInformation:   d.cmdGetInformation,
		// CmdGetPowerCurve: ,
		CmdSelectPowerCurve: d.cmdAckIgnored,
		CmdUpdatePowerCurve: d.cmdAckIgnored,
		CmdLog:              d.cmdAckIgnored,
	}

	for {
		frame, err := d.bus.ReadFrame()
		if err != nil {
			// Silently skip illegal frames, they are
			// typically ack frames from other devices.
			continue
		}

		if frame.Device != d.dev {
			// Skip frames that aren't targetting this
			// device
			continue
		}

		handler, ok := commandDispatchers[frame.Command]
		if !ok {
			fmt.Printf("Unexpected command: %v\n", frame.Command)
			continue
		}

		err = handler(frame)
		if err != nil {
			fmt.Printf("Failed to handle command %v: %v\n",
				frame.Command, err)
		}
	}
}

func (d *DeviceEmulator) sendAck(cmd *Frame) error {
	return d.bus.WriteAck(d.dev, cmd.Command)
}

func (d *DeviceEmulator) sendResp(req *Frame, resp []byte) error {
	frame := Frame{
		Device:  d.dev,
		Command: req.Command,
		Length:  uint8(len(resp)),
		Data:    resp,
	}

	return d.bus.WriteFrame(&frame)
}

func (d *DeviceEmulator) cmdAckIgnored(frame *Frame) error {
	return d.sendAck(frame)
}

func (d *DeviceEmulator) cmdGetInformation(frame *Frame) error {
	rdi, err := d.DeviceInformation.rawDeviceInfo()
	if err != nil {
		return err
	}

	resp, err := rdi.MarshalBinary()
	if err != nil {
		return err
	}

	return d.sendResp(frame, resp)
}
