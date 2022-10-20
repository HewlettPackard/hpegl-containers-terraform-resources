// (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
)

func DataSourceMachineBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:             schemas.MachineBlueprint(),
		ReadContext:        dataSourceMachineBlueprintReadContext,
		SchemaVersion:      0,
		StateUpgraders:     nil,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `Machine Blueprint data source allows reading machine blueprint data 
			based on blueprint name and appliance ID. Required inputs are name and site_id`,
	}
}

func dataSourceMachineBlueprintReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token: %s", err)
	}
	clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)

	var diags diag.Diagnostics

	applianceID := d.Get("site_id").(string)
	field := "applianceID eq " + applianceID
	blueprints, resp, err := c.CaasClient.ClusterAdminApi.V1MachineblueprintsGet(clientCtx, field, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var blueprint *mcaasapi.MachineBlueprint

	for b := range blueprints.Items {
		if blueprints.Items[b].Name == d.Get("name") {
			blueprint = &blueprints.Items[b]
			d.SetId(blueprint.Id)
		}
	}

	if blueprint == nil {
		return diag.Errorf("Machine blueprint '%s' not found in space '%s'", d.Get("name"), applianceID)
	}

	if err = writeMachineBlueprintResourceValues(d, blueprint); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
