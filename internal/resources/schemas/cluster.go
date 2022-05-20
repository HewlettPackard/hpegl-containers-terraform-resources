package schemas

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// nolint: funlen
func Cluster() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"state": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"health": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
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
			ForceNew: true,
		},
		"blueprint_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
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
		"machine_sets_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSetsDetail(),
			},
			ForceNew: true,
			Computed: true,
		},
		"api_endpoint": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
		"service_endpoints": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: ServiceEndpoints(),
			},
			ForceNew: true,
			Computed: true,
		},
		"site_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"appliance_name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"space_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"default_storage_class": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"default_storage_class_description": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
	}
}
