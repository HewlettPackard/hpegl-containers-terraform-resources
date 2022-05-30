// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package schemas

import (
	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func AvailableCapacity() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cpu": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"nodes": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"clusters": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
	}
}

func FlattenAvailableCapacity(availableCapacity *mcaasapi.ClusterProviderAvailableCapacity) []interface{} {
	if availableCapacity == nil {
		return nil
	}

	availablecapacity := make([]interface{}, 1)
	av := make(map[string]interface{})

	av["cpu"] = availableCapacity.Cpu
	av["nodes"] = availableCapacity.Nodes
	av["clusters"] = availableCapacity.Clusters
	availablecapacity[0] = av

	return availablecapacity
}
