// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP

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

func DataSourceClusterBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:             schemas.ClusterBlueprint(),
		ReadContext:        dataSourceClusterBlueprintReadContext,
		SchemaVersion:      0,
		StateUpgraders:     nil,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `Cluster Blueprint data source allows reading cluster blueprint data 
			based on blueprint name and space ID. Required inputs are name and space_id`,
	}
}

func dataSourceClusterBlueprintReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	siteID := d.Get("site_id").(string)

	blueprints, resp, err := c.CaasClient.ClusterAdminApi.V1ClusterblueprintsGet(clientCtx, siteID)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var blueprint *mcaasapi.ClusterBlueprint

	for b := range blueprints.Items {
		if blueprints.Items[b].Name == d.Get("name") {
			blueprint = &blueprints.Items[b]
			d.SetId(blueprint.Id)
		}
	}

	if blueprint == nil {
		return diag.Errorf("Cluster blueprint '%s' not found in site '%s'", d.Get("name"), siteID)
	}

	if err = writeBlueprintResourceValues(d, blueprint); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
