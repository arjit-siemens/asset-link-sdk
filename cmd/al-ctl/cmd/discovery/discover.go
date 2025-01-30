/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: MIT
 *
 */

package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/model"
	"os"

	"github.com/industrial-asset-hub/asset-link-sdk/v3/cmd/al-ctl/internal/al"
	"github.com/industrial-asset-hub/asset-link-sdk/v3/cmd/al-ctl/internal/shared"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var discoveryFile string = ""
var outputFile string = ""

// discoverCmd represents the discovery command
var DiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Start discovery job",
	Long:  `This command starts an discovery job and prints the result.`,
	Run: func(cmd *cobra.Command, args []string) {
		resp := al.Discover(shared.AssetLinkEndpoint, discoveryFile)
		for _, discoverResponse := range resp {
			for deviceIndex, discoveredDevice := range discoverResponse.Devices {
				transformedDevice := model.TransformDevice(discoveredDevice, "URI")
				log.Trace().Str("File", outputFile).Msg("Saving to file")
				f, _ := os.Create(fmt.Sprintf("test-device-%d.json", deviceIndex))
				asJson, _ := json.MarshalIndent(transformedDevice, "", "  ")
				_, err := f.Write(asJson)
				if err != nil {
					log.Err(err).Msg("error during writing of the json file")
				}
			}
		}
	},
}

func init() {
	DiscoverCmd.Flags().StringVarP(&outputFile, "output-file", "o", "result.json", "output file")
	DiscoverCmd.Flags().StringVarP(&discoveryFile, "discovery-file", "d", "", shared.DiscoveryFileDesc)
}
