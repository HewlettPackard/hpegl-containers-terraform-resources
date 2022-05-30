// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package schemas

import (
	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func StorageClasses() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"description": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"gl_storage_type": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"access_protocol": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"iops": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"encryption": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"dedupe": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"cost_per_gb": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
	}
}

func FlattenStorageClasses(storageClass *[]mcaasapi.StorageClass) []interface{} {
	if storageClass == nil {
		return nil
	}

	storageClasses := make([]interface{}, len(*storageClass))
	for i, sc := range *storageClass {
		sclasses := make(map[string]interface{})

		sclasses["name"] = sc.Name
		sclasses["description"] = sc.Description
		sclasses["gl_storage_type"] = sc.GlStorageType
		sclasses["access_protocol"] = sc.AccessProtocol
		sclasses["iops"] = sc.Iops
		sclasses["encryption"] = sc.Encryption
		sclasses["dedupe"] = sc.Dedupe
		sclasses["cost_per_gb"] = sc.CostPerGB
		storageClasses[i] = sclasses
	}

	return storageClasses
}
