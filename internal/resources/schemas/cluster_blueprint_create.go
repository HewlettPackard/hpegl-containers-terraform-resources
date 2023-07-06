package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterBlueprintCreate() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"kubernetes_version": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"default_storage_class": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"site_id": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"cluster_provider": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"control_plane_count": {
			Type:     schema.TypeFloat,
			ForceNew: true,
			Required: true,
		},
		"worker_nodes": {
			Type:     schema.TypeList,
			ForceNew: true,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"machine_blueprint_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"min_size": {
						Type:     schema.TypeFloat,
						Required: true,
					},
					"max_size": {
						Type:     schema.TypeFloat,
						Required: true,
					},
				},
			},
		},
	}
}
