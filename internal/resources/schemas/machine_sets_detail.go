package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
)

func MachineSetsDetail() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"os_image": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"os_version": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"min_size": {
			Type:     schema.TypeFloat,
			ForceNew: true,
			Computed: true,
		},
		"max_size": {
			Type:     schema.TypeFloat,
			ForceNew: true,
			Computed: true,
		},
		"machine_provider": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"machine_roles": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			ForceNew: true,
			Computed: true,
		},
		"compute_type": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"storage_type": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"size": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"size_detail": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: SizeDetail(),
			},
			ForceNew: true,
			Computed: true,
		},
		"networks": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			ForceNew: true,
			Computed: true,
		},
		"proxy": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"machines": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: Machines(),
			},
			ForceNew: true,
			Computed: true,
		},
	}
}

func FlattenMachineSetsDetail(machineSet *[]mcaasapi.MachineSetDetail) []interface{} {
	if machineSet == nil {
		return nil
	}

	machineSets := make([]interface{}, len(*machineSet))
	for i, machine := range *machineSet {
		mcset := make(map[string]interface{})

		mcset["name"] = machine.Name
		mcset["os_image"] = machine.OsImage
		mcset["os_version"] = machine.OsVersion
		mcset["min_size"] = machine.MinSize
		mcset["max_size"] = machine.MaxSize
		mcset["machine_provider"] = machine.MachineProvider
		mcset["machine_roles"] = machine.MachineRoles
		mcset["compute_type"] = machine.ComputeInstanceType
		mcset["storage_type"] = machine.StorageInstanceType
		mcset["size"] = machine.Size
		mcset["size_detail"] = FlattenSizeDetail(machine.SizeDetail)
		mcset["networks"] = machine.Networks
		mcset["machines"] = FlattenMachines(&machine.Machines)

		machineSets[i] = mcset
	}

	return machineSets
}
