// (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP

package resources

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/constants"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/registration"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources"
)

// Assert that Registration implements the ServiceRegistration interface
var _ registration.ServiceRegistration = (*Registration)(nil)

type Registration struct{}

func (r Registration) Name() string {
	return constants.ServiceName
}

func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"hpegl_caas_cluster_blueprint": resources.DataSourceClusterBlueprint(),
		"hpegl_caas_site":              resources.DataSourceAppliance(),
	}
}

func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"hpegl_caas_cluster_blueprint": resources.ClusterBlueprint(),
		"hpegl_caas_cluster":           resources.Cluster(),
	}
}

func (r Registration) ProviderSchemaEntry() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			constants.APIURL: {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HPEGL_CAAS_API_URL", ""),
				Description: "The URL to use for the CaaS API, can also be set with the HPEGL_CAAS_API_URL env var",
			},
		},
	}
}
