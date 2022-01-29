/*
 * SPDX-FileCopyrightText: Copyright 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"bytes"
	"reflect"
	"testing"
)

var (
	testDeviceInfoBinary = []byte{
		0xc2, 0x01, // Vin
		0x03, 0x02, // Iin
		0x74, 0x09, // Vgrid
		0x02, 0x01, // Igrid
		0xdc, 0x01, // Temp
		0xce, 0x04, 0x02, 0x01, // Tot. kWh
		0x11, 0x10, // State
		0xad, 0xde, // Err.
		0x96,       // Product
		0x0f,       // SW. version
		0x89, 0x13, // Grid freq.
		0x01,       // Power std.
		0x04,       // Power curve
		0x05, 0x04, // V2in
		0x06, 0x07, // I2in
		0xbe,       // Grid status
		0x2f, 0x04, // Month kWh
		0x8d, 0x03, // Last mth kWh
		0x15, 0x02, // Today kWh
		0x31, 0x01, // Yesterday
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, // Serial
		/*
			0x00, 0x00, 0x00, // Unknown
			0x01, 0x21, // Ext. ver?
		*/
	}

	testDeviceInfoRaw = rawDeviceInfo{
		VIn:             450,
		IIn:             515,
		VGrid:           2420,
		IGrid:           258,
		Temp:            476,
		TotalProduction: 16909518,
		Status:          0x1011,
		Error:           0xdead,
		Product:         0x96,
		SWVersion:       0xf,
		GridFreq:        5001,
		PowerStd:        0x1,
		PowerCurve:      0x4,
		V2In:            1029,
		I2In:            1798,
		GridStatus:      0xbe,
		MonthProd:       1071,
		LastMonthProd:   909,
		TodayProd:       533,
		YesterdayProd:   305,
		SerialNo:        [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
	}

	testDeviceInfo = DeviceInformation{
		Inputs: []InputStatus{
			InputStatus{Voltage: 45.0, Current: 51.5},
			InputStatus{Voltage: 102.9, Current: 179.8},
		},
		Grid: GridInformation{
			Voltage:       242.0,
			Current:       25.8,
			Frequency:     50.01,
			PowerStandard: PowerStandardG59G83,
			GridStatus:    GridStatus(0xbe),
		},
		Production: ProductionInformation{
			Total:     16909518,
			Month:     1071,
			LastMonth: 909,
			Today:     53.3,
			Yesterday: 30.5,
		},
		Temperature: 47.6,
		Product:     DeviceProduct(0x96),
		SWVersion:   DeviceVersion(0xf),
		SerialNo:    DeviceSerialNumber{1, 2, 3, 4, 5, 6, 7, 8},
		Status:      DeviceStatus(0x1011),
		Error:       DeviceError(0xdead),
		PowerCurve:  PowerCurve(0x04),
	}
)

func TestUnmarshalDeviceInformation(t *testing.T) {
	rdi := rawDeviceInfo{}
	rdi.UnmarshalBinary(testDeviceInfoBinary)

	if rdi != testDeviceInfoRaw {
		t.Error("Unmarshalled device info mismatch")
	}
}

func TestMarshalDeviceInformation(t *testing.T) {
	bin, err := testDeviceInfoRaw.MarshalBinary()

	if err != nil {
		t.Error("MarshalBinary failed: ", err)
	}

	if bytes.Compare(bin, testDeviceInfoBinary) != 0 {
		t.Error("Marshalled device info mismatch: ", bin, testDeviceInfoBinary)
	}
}

func TestDecodeDeviceInformation(t *testing.T) {
	di := testDeviceInfoRaw.DeviceInformation()
	if !reflect.DeepEqual(*di, testDeviceInfo) {
		t.Errorf("Decoded device info mismatch")
	}
}

func TestEncodeDeviceInformation(t *testing.T) {
	rdi, err := testDeviceInfo.rawDeviceInfo()
	if err != nil {
		t.Error("MarshalBinary failed: ", err)
	}

	if *rdi != testDeviceInfoRaw {
		t.Errorf("Encoded device info mismatch")
		t.Logf("%#v\n", rdi)
	}
}
