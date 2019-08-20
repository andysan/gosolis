/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func gridOnMain(cmd *cobra.Command, args []string) {
	dev := getInverter()

	fmt.Printf("Connecting to grid...\n")
	errComm(dev.GridOn())
	fmt.Printf("Done.\n")
}

func gridOffMain(cmd *cobra.Command, args []string) {
	dev := getInverter()

	fmt.Printf("Disconnecting from grid...\n")
	errComm(dev.GridOff())
	fmt.Printf("Done.\n")
}

var gridCmd = &cobra.Command{
	Use:   "grid",
	Short: "Grid control",
}

var gridOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Connect ot the grid",
	Args:  cobra.NoArgs,
	Run:   gridOnMain,
}

var gridOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disconnect the from the grid",
	Args:  cobra.NoArgs,
	Run:   gridOffMain,
}

func init() {
	RootCmd.AddCommand(gridCmd)
	gridCmd.AddCommand(gridOnCmd)
	gridCmd.AddCommand(gridOffCmd)
}
