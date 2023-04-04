// (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP

package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	libUtils "github.com/hewlettpackard/hpegl-provider-lib/pkg/utils"

	testutils "github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/test-utils"
)

func main() {
	// plugin.Serve(&plugin.ServeOpts{
	// 	ProviderFunc: testutils.ProviderFunc(),
	// })

	// Read config file for acceptance test if TF_ACC sets
	libUtils.ReadAccConfig(".")

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	opts := &plugin.ServeOpts{
		ProviderFunc: testutils.ProviderFunc(),
	}
	if debugMode {
		optsDebug := &plugin.ServeOpts{
			ProviderFunc: testutils.ProviderFunc(),
			Debug:        true,
		}
		plugin.Serve(optsDebug)

		return
	}

	plugin.Serve(opts)
}
