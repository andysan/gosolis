/*
 * SPDX-FileCopyrightText: Copyright 2019, 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

// Illegal response received. This can be caused by an unexpected
// device ID or command.
var IllegalResponseError = errors.New("Illegal response")

type Device struct {
	bus BusInterface
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
	Voltage float64
	Current float64
}

type GridInformation struct {
	Voltage       float64
	Current       float64
	Frequency     float64
	PowerStandard PowerStandard
	GridStatus    GridStatus
}

type ProductionInformation struct {
	// Lifetime production (kWh)
	Total float64
	// Production this month  (kWh)
	Month float64
	// Production last month  (kWh)
	LastMonth float64
	// Production today  (kWh)
	Today float64
	// Production yesterday  (kWh)
	Yesterday float64
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

type DeviceStatus uint16

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
	Temperature float64

	Product   DeviceProduct
	SWVersion DeviceVersion
	SerialNo  DeviceSerialNumber
	Status    DeviceStatus
	Error     DeviceError

	PowerCurve PowerCurve
}

// Instantiate a Solis device interface
func NewDevice(bus BusInterface, dev DeviceId) *Device {
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

func (rpi *rawDeviceInfo) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, rpi)

	return buf.Bytes(), nil
}

func (di *DeviceInformation) rawDeviceInfo() (*rawDeviceInfo, error) {
	if len(di.Inputs) > 2 {
		return nil, IllegalResponseError
	}

	rdi := rawDeviceInfo{
		VGrid:           uint16(math.Round(di.Grid.Voltage * 10)),
		IGrid:           uint16(math.Round(di.Grid.Current * 10)),
		Temp:            uint16(math.Round(di.Temperature * 10)),
		TotalProduction: uint32(math.Round(di.Production.Total)),
		Status:          uint16(di.Status),
		Error:           uint16(di.Error),
		Product:         uint8(di.Product),
		SWVersion:       uint8(di.SWVersion),
		GridFreq:        uint16(math.Round(di.Grid.Frequency * 100)),
		PowerStd:        uint8(di.Grid.PowerStandard),
		PowerCurve:      uint8(di.PowerCurve),
		GridStatus:      uint8(di.Grid.GridStatus),
		MonthProd:       uint16(math.Round(di.Production.Month)),
		LastMonthProd:   uint16(math.Round(di.Production.LastMonth)),
		TodayProd:       uint16(math.Round(di.Production.Today * 10)),
		YesterdayProd:   uint16(math.Round(di.Production.Yesterday * 10)),
		SerialNo:        [8]byte(di.SerialNo),
	}

	if len(di.Inputs) >= 1 {
		rdi.VIn = uint16(math.Round(di.Inputs[0].Voltage * 10))
		rdi.IIn = uint16(math.Round(di.Inputs[0].Current * 10))
	}

	if len(di.Inputs) >= 2 {
		rdi.V2In = uint16(math.Round(di.Inputs[1].Voltage * 10))
		rdi.I2In = uint16(math.Round(di.Inputs[1].Current * 10))
	}

	return &rdi, nil
}

func (rpi *rawDeviceInfo) DeviceInformation() *DeviceInformation {
	pi := DeviceInformation{
		Inputs: []InputStatus{
			InputStatus{
				Voltage: float64(rpi.VIn) / 10.0,
				Current: float64(rpi.IIn) / 10.0,
			},
			InputStatus{
				Voltage: float64(rpi.V2In) / 10.0,
				Current: float64(rpi.I2In) / 10.0,
			},
		},
		Grid: GridInformation{
			Voltage:       float64(rpi.VGrid) / 10.0,
			Current:       float64(rpi.IGrid) / 10.0,
			Frequency:     float64(rpi.GridFreq) / 100.0,
			PowerStandard: PowerStandard(rpi.PowerStd),
			GridStatus:    GridStatus(rpi.GridStatus),
		},
		Production: ProductionInformation{
			Total:     float64(rpi.TotalProduction),
			Month:     float64(rpi.MonthProd),
			LastMonth: float64(rpi.LastMonthProd),
			Today:     float64(rpi.TodayProd) / 10.0,
			Yesterday: float64(rpi.YesterdayProd) / 10.0,
		},
		Temperature: float64(rpi.Temp) / 10.0,
		Product:     DeviceProduct(rpi.Product),
		SWVersion:   DeviceVersion(rpi.SWVersion),
		SerialNo:    DeviceSerialNumber(rpi.SerialNo),
		Status:      DeviceStatus(rpi.Status),
		Error:       DeviceError(rpi.Error),
		PowerCurve:  PowerCurve(rpi.PowerCurve),
	}

	return &pi
}
