// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package testutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/provider"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/common"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/retrieve"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/serviceclient"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/resources"
)

func ProviderFunc() plugin.ProviderFunc {
	return provider.NewProviderFunc(provider.ServiceRegistrationSlice(resources.Registration{}), providerConfigure)
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc { // nolint staticcheck
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cli, err := client.InitialiseClient{}.NewClient(d)
		if err != nil {
			return nil, diag.Errorf("error in creating client: %s", err)
		}
		// Initialise token handler
		h, err := serviceclient.NewHandler(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Returning a map[string]interface{} with the Client from pkg.client at the
		// key specified in that repo and with the token retrieve function at the key
		// specified by the token package to ensure compatibility with the hpegl terraform
		// provider.
		return map[string]interface{}{
			client.InitialiseClient{}.ServiceName(): cli,
			common.TokenRetrieveFunctionKey:         retrieve.NewTokenRetrieveFunc(h),
		}, nil
	}
}
