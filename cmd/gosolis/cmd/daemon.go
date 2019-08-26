/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"log"
	"time"

	solis "github.com/andysan/gosolis/pkg/gosolis"
	"github.com/andysan/gosolis/pkg/hermes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func daemonSendReport(bus *hermes.Hermes, di *solis.DeviceInformation) {
	msg := map[string]interface{}{
		"v_in":       di.Inputs[0].Voltage,
		"i_in":       di.Inputs[0].Current,
		"v_grid":     di.Grid.Voltage,
		"i_grid":     di.Grid.Current,
		"f_grid":     di.Grid.Frequency,
		"temp":       di.Temperature,
		"production": di.Production.Total,
	}
	if err := bus.Send(msg); err != nil {
		log.Println("Message bus send failed: ", err)
	}
}

func daemonMain(cmd *cobra.Command, args []string) {
	h := viper.Sub("hermes")
	if h == nil {
		log.Fatal("Hermes not configured")
	}

	dev := getInverter()

	bus := hermes.NewViper(h)
	if bus == nil {
		log.Fatal("Failed to create Hermes backend")
	}

	for {
		if di, err := dev.GetInformation(); err == nil {
			daemonSendReport(bus, di)
		} else {
			log.Println("Failed to get device report: ", err)
		}

		time.Sleep(config.Daemon.Interval)
	}
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Solis logging daemon",
	Args:  cobra.NoArgs,
	Run:   daemonMain,
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}
