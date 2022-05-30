// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package schemas

import (
	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func LicenseInfo() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"label": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"summary": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			ForceNew: true,
			Computed: true,
		},
	}
}

func FlattenLicenseInfo(licenseInfo *mcaasapi.ClusterProviderLicenseInfo) []interface{} {
	if licenseInfo == nil {
		return nil
	}
	lic := licenseInfo.Licenses
	licenses := make([]interface{}, len(lic))
	for i, lf := range lic {
		licinfo := make(map[string]interface{})

		licinfo["label"] = lf.Label
		licinfo["summary"] = lf.Summary
		licinfo["status"] = lf.Status
		licenses[i] = licinfo
	}

	return licenses
}
