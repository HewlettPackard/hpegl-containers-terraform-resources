// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterProvider() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"site_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"created_date": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"last_update_date": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"id": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"state": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"health": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"kubernetes_versions": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			ForceNew: true,
			Computed: true,
		},
		"min_master_size": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: SizeDetail(),
			},
			ForceNew: true,
			Computed: true,
		},
		"min_worker_size": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: SizeDetail(),
			},
			ForceNew: true,
			Computed: true,
		},
		"license_info": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: LicenseInfo(),
			},
			ForceNew: true,
			Computed: true,
		},
		"storage_classes": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: StorageClasses(),
			},
			ForceNew: true,
			Computed: true,
		},
		"available_capacity": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: AvailableCapacity(),
			},
			ForceNew: true,
			Computed: true,
		},
	}
}
