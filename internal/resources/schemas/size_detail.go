package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
)

func SizeDetail() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"cpu": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"root_disk": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"ephemeral_disk": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
		"persistent_disk": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Computed: true,
		},
	}
}

func FlattenSizeDetail(sizeDetail *mcaasapi.AllOfMachineSetDetailSizeDetail) []interface{} {
	if sizeDetail == nil {
		return nil
	}

	sizesOut := make([]interface{}, 1)
	sizeOut := make(map[string]interface{})

	sizeOut["name"] = sizeDetail.Name
	sizeOut["cpu"] = sizeDetail.Cpu
	sizeOut["memory"] = sizeDetail.Memory
	sizeOut["root_disk"] = sizeDetail.RootDisk
	sizeOut["ephemeral_disk"] = sizeDetail.EphemeralDisk
	sizeOut["persistent_disk"] = sizeDetail.PersistentDisk
	sizesOut[0] = sizeOut

	return sizesOut
}

func FlattenSizeDetailMachineBlueprint(sizeDetail *mcaasapi.AllOfMachineBlueprintSizeDetail) []interface{} {
	if sizeDetail == nil {
		return nil
	}

	sizesOut := make([]interface{}, 1)
	sizeOut := make(map[string]interface{})

	sizeOut["name"] = sizeDetail.Name
	sizeOut["cpu"] = sizeDetail.Cpu
	sizeOut["memory"] = sizeDetail.Memory
	sizeOut["root_disk"] = sizeDetail.RootDisk
	sizeOut["ephemeral_disk"] = sizeDetail.EphemeralDisk
	sizeOut["persistent_disk"] = sizeDetail.PersistentDisk
	sizesOut[0] = sizeOut

	return sizesOut
}
