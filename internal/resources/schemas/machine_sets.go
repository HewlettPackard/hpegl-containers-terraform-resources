package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
)

func MachineSets() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"machine_blueprint_id": {
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
	}
}

func FlattenMachineSets(machineSet *[]mcaasapi.MachineSet) []interface{} {
	if machineSet == nil {
		return nil
	}

	machineSets := make([]interface{}, len(*machineSet))
	for i, machine := range *machineSet {
		mcset := make(map[string]interface{})

		mcset["name"] = machine.Name
		mcset["machine_blueprint_id"] = machine.MachineBlueprintId
		mcset["min_size"] = machine.MinSize
		mcset["max_size"] = machine.MaxSize
		machineSets[i] = mcset
	}

	return machineSets
}
