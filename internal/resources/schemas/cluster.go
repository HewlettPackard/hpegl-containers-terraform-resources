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
			Computed: true,
		},
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
			Required: true,
		},
		"blueprint_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"kubernetes_version": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"cluster_provider": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"default_machine_sets": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSets(),
			},
			Computed: true,
		},
		"default_machine_sets_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSetsDetail(),
			},
			Computed: true,
		},
		"machine_sets": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSets(),
			},
			Computed: true,
		},
		"machine_sets_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSetsDetail(),
			},
			Computed: true,
		},
		"api_endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"service_endpoints": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: ServiceEndpoints(),
			},
			Computed: true,
		},
		"site_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"appliance_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"space_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"default_storage_class": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"default_storage_class_description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"kubeconfig": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"worker_nodes": {
			Type:     schema.TypeList,
			Optional: true,
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
					"count": {
						Type:     schema.TypeFloat,
						Required: true,
					},
					"os_version": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"os_image": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

func DataCluster() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"space_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"state": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"health": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"created_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_update_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"blueprint_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"kubernetes_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"cluster_provider": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"machine_sets": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSets(),
			},
			Computed: true,
		},
		"machine_sets_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: MachineSetsDetail(),
			},
			Computed: true,
		},
		"api_endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"service_endpoints": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: ServiceEndpoints(),
			},
			Computed: true,
		},
		"site_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"appliance_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"default_storage_class": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"default_storage_class_description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"kubeconfig": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
