/*
 * SPDX-FileCopyrightText: Copyright 2019, 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"fmt"
	"os"
	"time"

	solis "github.com/andysan/gosolis/pkg/gosolis"
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
	Type    string
	Port    string
	Addr    solis.DeviceId
	Baud    uint
	Timeout time.Duration
}

type DaemonConfig struct {
	Interval      time.Duration
	ProbeInterval time.Duration `mapstructure:"probe_interval"`
}

type Config struct {
	Inverter InverterConfig
	Daemon   DaemonConfig
}

var (
	cfgFile string
	config  Config
)

var (
	solisBus    solis.BusInterface
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

func createBusSerial() solis.BusInterface {
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

	return solis.NewSerialBus(trw)
}

func getBus() solis.BusInterface {
	if solisBus != nil {
		return solisBus
	}

	switch config.Inverter.Type {
	case "serial":
		solisBus = createBusSerial()
	default:
		fmt.Printf("Incorrect bus type: %s\n", config.Inverter.Type)
	}

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
		"config file (default is /etc/gosolis.{yaml,json,...})")
	RootCmd.PersistentFlags().StringP(
		"bus-type", "b", "serial",
		"device type (serial or test)")
	RootCmd.PersistentFlags().StringP(
		"port", "p", "",
		"serial interface connected to inverter(s)")
	RootCmd.PersistentFlags().IntP(
		"addr", "a", 1,
		"Inverter ID on bus")
	RootCmd.PersistentFlags().IntP(
		"timeout", "t", 100,
		"Timeout in milliseconds")

	errPanic(viper.BindPFlag("inverter.type", pfs.Lookup("bus-type")))
	errPanic(viper.BindPFlag("inverter.port", pfs.Lookup("port")))
	errPanic(viper.BindPFlag("inverter.addr", pfs.Lookup("addr")))
	errPanic(viper.BindPFlag("inverter.timeout", pfs.Lookup("timeout")))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc")
		viper.SetConfigName("gosolis")
	}

	viper.SetDefault("inverter.type", "serial")
	viper.SetDefault("inverter.port", "")
	viper.SetDefault("inverter.addr", 1)
	viper.SetDefault("inverter.baud", 9600)
	viper.SetDefault("inverter.timeout", 500*time.Millisecond)

	viper.SetDefault("daemon.interval", 10*time.Second)
	viper.SetDefault("daemon.probe_interval", 1*time.Minute)

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
