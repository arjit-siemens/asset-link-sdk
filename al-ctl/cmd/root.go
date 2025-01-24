/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: MIT
 *
 */

package cmd

import (
	"fmt"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/al-ctl/shared"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/al-ctl/testsuite"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/logging"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

var (
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "al-ctl",
	Short: "command line interface to interact with Asset Links",
	Long: `This command line interfaces allows to interact with the so called Asset Links (AL).

This can be useful for validation purposes inside CI/CD pipelines or just
to ease development efforts.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initHandlers)
	rootCmd.PersistentFlags().StringVarP(&shared.RegistryEndpoint, "registry", "r", "localhost:50051", "gRPC Server Address of the Registry")
	rootCmd.PersistentFlags().StringVarP(&shared.AssetLinkEndpoint, "endpoint", "e", "localhost:8081", "gRPC Server Address of the AssetLink")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "info",
		fmt.Sprintf("set log level. one of: %s,%s,%s,%s,%s,%s,%s",
			zerolog.TraceLevel.String(),
			zerolog.DebugLevel.String(),
			zerolog.InfoLevel.String(),
			zerolog.WarnLevel.String(),
			zerolog.ErrorLevel.String(),
			zerolog.FatalLevel.String(),
			zerolog.PanicLevel.String()))

	rootCmd.AddCommand(DiscoverCmd)
	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(testsuite.TestCmd)

}
func initHandlers() {
	logging.SetupLogging()
	logging.AdjustLogLevel(logLevel)
}
