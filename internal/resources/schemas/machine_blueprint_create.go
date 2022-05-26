package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func MachineBlueprintCreate() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"created_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_update_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"machine_provider": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"machine_roles": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			ForceNew: true,
			Required: true,
		},
		"os_image": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"os_version": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"size": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"size_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: SizeDetail(),
			},
			ForceNew: true,
			Computed: true,
		},
		"compute_type": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"storage_type": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"site_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}
