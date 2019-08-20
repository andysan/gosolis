/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"fmt"
	"os"
	"time"

	solis "github.com/andysan/gosolis/pkg/gosolis"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tarm/serial"
)

const (
	exitUsage  = 1
	exitConfig = 2
	exitSerial = 3
)

type InverterConfig struct {
	Port    string
	Addr    solis.DeviceId
	Baud    uint
	Timeout time.Duration
}

type Config struct {
	Inverter InverterConfig
}

var (
	cfgFile string
	config  Config
)

var (
	solisBus    *solis.Bus
	solisDevice *solis.Device
)

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func errComm(err error) {
	if err != nil {
		fmt.Println("Communication error:", err)
		os.Exit(exitSerial)
	}
}

func getBus() *solis.Bus {
	if solisBus != nil {
		return solisBus
	}

	if config.Inverter.Port == "" {
		fmt.Println("No serial port specified")
		os.Exit(exitUsage)
	}

	c := &serial.Config{
		Name: config.Inverter.Port,
		Baud: int(config.Inverter.Baud),
	}

	port, err := serial.OpenPort(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(exitSerial)
	}

	timeout := config.Inverter.Timeout
	trw := solis.NewTimeoutReadWriter(port, timeout, 16)

	solisBus = solis.NewBus(trw)

	return solisBus
}

func getInverter() *solis.Device {
	if solisDevice != nil {
		return solisDevice
	}

	bus := getBus()
	solisDevice := solis.NewDevice(bus, config.Inverter.Addr)

	fmt.Println("Connecting to inverter...")
	if err := solisDevice.Ping(); err != nil {
		fmt.Println("Failed to connect to inverter:", err)
		os.Exit(exitSerial)
	}

	fmt.Println("Found inverter...")
	return solisDevice
}

func rootPersistentPreRun(cmd *cobra.Command, args []string) {
	viper.Unmarshal(&config)
}

var RootCmd = &cobra.Command{
	Use:              "gosolis",
	Short:            "Ginlong Solis control and monitor",
	Long:             `gosolis is a control and monitor application for Ginlong Solis inverters.`,
	PersistentPreRun: rootPersistentPreRun,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(exitUsage)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	pfs := RootCmd.PersistentFlags()
	pfs.StringVarP(
		&cfgFile, "config", "c", "",
		"config file (default is $HOME/.gosolis.yaml)")
	RootCmd.PersistentFlags().StringP(
		"port", "p", "",
		"serial interface connected to inverter(s)")
	RootCmd.PersistentFlags().IntP(
		"addr", "a", 1,
		"Inverter ID on bus")
	RootCmd.PersistentFlags().IntP(
		"timeout", "t", 100,
		"Timeout in milliseconds")

	errPanic(viper.BindPFlag("inverter.port", pfs.Lookup("port")))
	errPanic(viper.BindPFlag("inverter.addr", pfs.Lookup("addr")))
	errPanic(viper.BindPFlag("inverter.timeout", pfs.Lookup("timeout")))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(exitConfig)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".gosolis")
	}

	viper.SetDefault("inverter.port", "")
	viper.SetDefault("inverter.addr", 1)
	viper.SetDefault("inverter.baud", 9600)
	viper.SetDefault("inverter.timeout", 500*time.Millisecond)

	viper.SetEnvPrefix("gosolis")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("Failed to read config:", err)
			os.Exit(exitConfig)
		}
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
