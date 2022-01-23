/*
 * SPDX-FileCopyrightText: Copyright 2019 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Illegal response received. This can be caused by an unexpected
// device ID or command.
var IllegalResponseError = errors.New("Illegal response")

type Device struct {
	bus *Bus
	dev DeviceId
}

type PowerStandard uint8

const (
	// ?
	PowerStandardDefault  = PowerStandard(0x00)
	PowerStandardG59G83   = PowerStandard(0x01)
	PowerStandardUL240V   = PowerStandard(0x02)
	PowerStandardVDE0126  = PowerStandard(0x03)
	PowerStandardAS4777   = PowerStandard(0x04)
	PowerStandardAS4777NQ = PowerStandard(0x05)
	PowerStandardCQC      = PowerStandard(0x06)
	PowerStandardENEL     = PowerStandard(0x07)
	PowerStandardUL208V   = PowerStandard(0x08)
	PowerStandardMEXCFE   = PowerStandard(0x09)
	// User defined
	PowerStandardUser      = PowerStandard(0x0A)
	PowerStandardVDE4105   = PowerStandard(0x0B)
	PowerStandardEN50438DK = PowerStandard(0x0C)
	PowerStandardEN50438IE = PowerStandard(0x0D)
	PowerStandardEN50438NL = PowerStandard(0x0E)
	PowerStandardEN50438T  = PowerStandard(0x0F)
	PowerStandardEN50438L  = PowerStandard(0x10)
)

type GridStatus uint8

const ()

type InputStatus struct {
	Voltage float32
	Current float32
}

type GridInformation struct {
	Voltage       float32
	Current       float32
	Frequency     float32
	PowerStandard PowerStandard
	GridStatus    GridStatus
}

type ProductionInformation struct {
	// Lifetime production (kWh)
	Total float32
	// Production this month  (kWh)
	Month float32
	// Production last month  (kWh)
	LastMonth float32
	// Production today  (kWh)
	Today float32
	// Production yesterday  (kWh)
	Yesterday float32
}

type rawDeviceInfo struct {
	VIn             uint16
	IIn             uint16
	VGrid           uint16
	IGrid           uint16
	Temp            uint16
	TotalProduction uint32
	Status          uint16
	Error           uint16
	Product         uint8
	SWVersion       uint8
	GridFreq        uint16
	PowerStd        uint8
	PowerCurve      uint8
	V2In            uint16
	I2In            uint16
	GridStatus      uint8
	MonthProd       uint16
	LastMonthProd   uint16
	TodayProd       uint16
	YesterdayProd   uint16
	SerialNo        [8]byte
}

type DeviceStatus uint8

const ()

type DeviceError uint16

const ()

type DeviceProduct uint8

const ()

type DeviceVersion uint8

type PowerCurve uint8

type DeviceSerialNumber [8]byte

type DeviceInformation struct {
	Inputs     []InputStatus
	Grid       GridInformation
	Production ProductionInformation

	// Temperature (°C)
	Temperature float32

	Product   DeviceProduct
	SWVersion DeviceVersion
	SerialNo  DeviceSerialNumber
	Status    DeviceStatus
	Error     DeviceError

	PowerCurve PowerCurve
}

// Instantiate a Solis device interface
func NewDevice(bus *Bus, dev DeviceId) *Device {
	return &Device{bus, dev}
}

func (d *Device) verifyResponse(f *Frame, cmd Command) (*Frame, error) {
	if f.Device != d.dev || f.Command != cmd {
		return f, IllegalResponseError
	} else {
		return f, nil
	}
}

func (d *Device) waitForAck(cmd Command) (*Frame, error) {
	if f, e := d.bus.ReadAckFrame(); e != nil {
		return f, e
	} else {
		return d.verifyResponse(f, cmd)
	}
}

func (d *Device) waitForResponse(cmd Command) (*Frame, error) {
	if f, e := d.bus.ReadFrame(); e != nil {
		return f, e
	} else {
		return d.verifyResponse(f, cmd)
	}
}

func (d *Device) sendAckedCommand(cmd Command, data []byte) error {
	f := Frame{d.dev, cmd, uint8(len(data)), data}
	if e := d.bus.WriteFrame(&f); e != nil {
		return e
	}

	if _, e := d.waitForAck(cmd); e != nil {
		return e
	}

	return nil
}

func (d *Device) sendCommand(cmd Command, data []byte) (*Frame, error) {
	f := Frame{d.dev, cmd, uint8(len(data)), data}
	if e := d.bus.WriteFrame(&f); e != nil {
		return nil, e
	}

	return d.waitForResponse(cmd)
}

func (d *Device) Ping() error {
	return d.sendAckedCommand(CmdPing, nil)
}

func (d *Device) GridOn() error {
	return d.sendAckedCommand(CmdGridOn, nil)
}

func (d *Device) GridOff() error {
	return d.sendAckedCommand(CmdGridOff, nil)
}

func (d *Device) GetInformation() (*DeviceInformation, error) {
	f, e := d.sendCommand(CmdGetInformation, nil)
	if e != nil {
		return nil, e
	}

	rdi := rawDeviceInfo{}
	if e := rdi.UnmarshalBinary(f.Data); e != nil {
		return nil, e
	}

	return rdi.DeviceInformation(), nil
}

func (rpi *rawDeviceInfo) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	return binary.Read(r, binary.LittleEndian, rpi)
}

func (rpi *rawDeviceInfo) DeviceInformation() *DeviceInformation {
	pi := DeviceInformation{
		Inputs: []InputStatus{
			InputStatus{
				Voltage: float32(rpi.VIn) / 10.0,
				Current: float32(rpi.IIn) / 10.0,
			},
			InputStatus{
				Voltage: float32(rpi.V2In) / 10.0,
				Current: float32(rpi.I2In) / 10.0,
			},
		},
		Grid: GridInformation{
			Voltage:       float32(rpi.VGrid) / 10.0,
			Current:       float32(rpi.IGrid) / 10.0,
			Frequency:     float32(rpi.GridFreq) / 100.0,
			PowerStandard: PowerStandard(rpi.PowerStd),
			GridStatus:    GridStatus(rpi.GridStatus),
		},
		Production: ProductionInformation{
			Total:     float32(rpi.TotalProduction),
			Month:     float32(rpi.MonthProd),
			LastMonth: float32(rpi.LastMonthProd),
			Today:     float32(rpi.TodayProd) / 10.0,
			Yesterday: float32(rpi.YesterdayProd) / 10.0,
		},
		Temperature: float32(rpi.Temp) / 10.0,
		Product:     DeviceProduct(rpi.Product),
		SWVersion:   DeviceVersion(rpi.SWVersion),
		SerialNo:    DeviceSerialNumber(rpi.SerialNo),
		Status:      DeviceStatus(rpi.Status),
		Error:       DeviceError(rpi.Error),
		PowerCurve:  PowerCurve(rpi.PowerCurve),
	}

	return &pi
}
