package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterBlueprint() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"site_id": {
			Type:     schema.TypeString,
			Required: true,
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
			Required: true,
		},
		"k8s_version": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"cluster_provider": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"machine_sets": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSets(),
			},
			ForceNew: true,
			Computed: true,
		},
		"default_storage_class": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
	}
}
