/*
 * SPDX-FileCopyrightText: Copyright 2019 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func statusMain(cmd *cobra.Command, args []string) {
	dev := getInverter()

	di, err := dev.GetInformation()
	errComm(err)

	for idx, inp := range di.Inputs {
		fmt.Printf("Input %d: %.1f V / %.1f A\n",
			idx, inp.Voltage, inp.Current)
	}

	fmt.Printf("Grid: %.1f V / %.1f A @ %.2f Hz\n",
		di.Grid.Voltage, di.Grid.Current, di.Grid.Frequency)

	p := &di.Production
	fmt.Printf("Production:\n")
	fmt.Printf("\tTotal: %.0f\n", p.Total)
	fmt.Printf("\tToday: %.1f\n", p.Today)
	fmt.Printf("\tYesterday: %.1f\n", p.Yesterday)
	fmt.Printf("\tThis month: %.0f\n", p.Month)
	fmt.Printf("\tLast month: %.0f\n", p.LastMonth)

	fmt.Printf("Inverter:\n")
	fmt.Printf("\tProduct type: %#x\n", di.Product)
	fmt.Printf("\tSoftware version: %#x\n", di.SWVersion)
	fmt.Printf("\tSerial: %#x\n", di.SerialNo)
	fmt.Printf("\tTemperature: %.1f °C\n", di.Temperature)
	fmt.Printf("\tStatus: %v\n", di.Status)
	fmt.Printf("\tError code: %#.4x\n", di.Error)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get inverter status",
	Args:  cobra.NoArgs,
	Run:   statusMain,
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
