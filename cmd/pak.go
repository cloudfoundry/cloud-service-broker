// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/brokerpak"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	pakCachePath = "pak.cache_path"
)

func init() {
	_ = viper.BindEnv(pakCachePath, "PAK_BUILD_CACHE_PATH")

	pakGroup := &cobra.Group{
		ID:    "pak",
		Title: "Brokerpak Development",
	}

	rootCmd.AddGroup(pakGroup)

	pakCmd := &cobra.Command{
		Use:     "pak",
		GroupID: "pak",
		Short:   "interact with user-defined service definition bundles",
		Long: `Lets you create, validate, and view service definition bundles.

A service definition bundle is a zip file containing all the elements needed
to define and run a custom service.

Bundles include source code (for legal compliance), service definitions, and
OpenTofu/provider binaries for multiple platforms. They give you a contained
way to deploy new services to existing brokers or augment the broker to fit
your needs.

To start building a pack, create a new directory and within it run init:

	cloud-service-broker pak init my-pak

You'll get a new pack with a manifest and example service definition.
Define the architectures and OpenTofu plugins you need in your manifest along
with any metadata you want, and include the names of all service definition
files.

When you're done, you can build the bundle which will download the sources,
OpenTofu resources, and pack them together.

	cloud-service-broker pak build my-pak

This will produce a pack:

	my-pak.brokerpak

You can run validation on an existing pack you created or downloaded:

	cloud-service-broker pak validate my-pak.brokerpak

You can also list information about the pack which includes metadata,
dependencies, services it provides, and the contents.

	cloud-service-broker pak info my-pak.brokerpak

`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	rootCmd.AddCommand(pakCmd)

	pakCmd.AddCommand(&cobra.Command{
		Use:   "init [path/to/pack/directory]",
		Short: "initialize a brokerpak manifest and example service",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			directory := ""
			if len(args) == 1 {
				directory = args[0]
			}

			if err := brokerpak.Init(directory); err != nil {
				log.Fatalf("error while packing %q: %v", directory, err)
			}
		},
	})

	const (
		includeSourceFlag = "include-source"
		targetFlag        = "target"
		compressFlag      = "compress"
	)
	buildCmd := &cobra.Command{
		Use:   "build [path/to/pack/directory]",
		Short: "bundle up the service definition files and OpenTofu resources into a brokerpak",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			directory := ""
			if len(args) == 1 {
				directory = args[0]
			}

			includeSource, err := cmd.Flags().GetBool(includeSourceFlag)
			if err != nil {
				log.Fatalf("error while obtaining the %q flag: %s", includeSourceFlag, err)
			}
			compress, err := cmd.Flags().GetBool(compressFlag)
			if err != nil {
				log.Fatalf("error while obtaining the %q flag: %s", compressFlag, err)
			}
			target, err := cmd.Flags().GetString(targetFlag)
			if err != nil {
				log.Fatalf("error while obtaining the %q flag: %s", targetFlag, err)
			}

			pakPath, err := brokerpak.Pack(directory, viper.GetString(pakCachePath), includeSource, compress, platform.Parse(target))
			if err != nil {
				log.Fatalf("error while packing %q: %v", directory, err)
			}

			if err := brokerpak.Validate(pakPath); err != nil {
				log.Fatalf("created: %v, but it failed validity checking: %v\n", pakPath, err)
			} else {
				fmt.Printf("created: %v\n", pakPath)
			}
		},
	}
	buildCmd.Flags().BoolP(includeSourceFlag, "s", false, "include source in the brokerpak")
	buildCmd.Flags().Bool(compressFlag, true, "compress the brokerpak")
	buildCmd.Flags().StringP(targetFlag, "t", "", "target specified platform; format 'darwin/amd64'; or special case 'current'")
	pakCmd.AddCommand(buildCmd)

	pakCmd.AddCommand(&cobra.Command{
		Use:   "info [pack.brokerpak]",
		Short: "get info about a brokerpak",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := brokerpak.Info(args[0]); err != nil {
				log.Fatalf("error getting info for %q: %v", args[0], err)
			}
		},
	})

	pakCmd.AddCommand(&cobra.Command{
		Use:   "validate [pack.brokerpak]",
		Short: "validate a brokerpak",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := brokerpak.Validate(args[0]); err != nil {
				log.Fatalf("Error: %v\n", err)
			} else {
				log.Println("Valid")
			}
		},
	})

	pakCmd.AddCommand(&cobra.Command{
		Use:   "run-examples [pack.brokerpak]",
		Short: "run the examples from a brokerpak",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			brokerpak.RunExamples(args[0])
		},
	})

	pakCmd.AddCommand(&cobra.Command{
		Use:     "docs [pack.brokerpak]",
		Aliases: []string{"use"},
		Short:   "generate the markdown usage docs for the given pack",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_ = brokerpak.Docs(args[0])
		},
	})

	pakCmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Run an integration test for the workflow",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Runs a quick and dirty e2e test for the development pattern
			td, err := os.MkdirTemp("", "test-brokerpak")
			if err != nil {
				log.Fatalf("couldn't initialize temp directory: %v", err)
			}
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(td)

			if err := brokerpak.Init(td); err != nil {
				log.Fatalf("couldn't initialize brokerpak: %v", err)
			}

			// Edit the manifest to point to our local server
			packname, err := brokerpak.Pack(td, "", false, true, platform.Platform{})
			defer os.Remove(packname)
			if err != nil {
				log.Fatalf("couldn't pack brokerpak: %v", err)
			}

			if err := brokerpak.Validate(packname); err != nil {
				log.Fatalf("couldn't validate brokerpak: %v", err)
			}

			log.Println("success!")
		},
	})
}
