/*
 * SPDX-FileCopyrightText: Copyright 2019 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configSaveMain(cmd *cobra.Command, args []string) {
	viper.WriteConfigAs(args[0])
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration",
}

var configSaveCmd = &cobra.Command{
	Use:   "save FILE",
	Short: "Save configuration to file",
	Args:  cobra.ExactArgs(1),
	Run:   configSaveMain,
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSaveCmd)
}
