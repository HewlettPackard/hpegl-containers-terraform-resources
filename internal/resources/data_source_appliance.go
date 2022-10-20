// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

func DataSourceAppliance() *schema.Resource {
	return &schema.Resource{
		Schema:             schemas.Appliance(),
		ReadContext:        dataSourceApplianceReadContext,
		SchemaVersion:      0,
		StateUpgraders:     nil,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `Appliance data source allows reading appliance data 
			based on name and space ID. Required inputs are name and space_id`,
	}
}

func dataSourceApplianceReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	spaceID := d.Get("space_id").(string)
	field := "spaceID eq " + spaceID
	appliances, resp, err := c.CaasClient.SiteApi.V1AppliancesGet(clientCtx, field, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var appliance *mcaasapi.Appliance

	for b := range appliances.Items {
		if appliances.Items[b].Name == d.Get("name") {
			appliance = &appliances.Items[b]
			d.SetId(appliance.Id)
		}
	}

	if appliance == nil {
		return diag.Errorf("Appliance '%s' not found in space '%s'", d.Get("name"), spaceID)
	}

	if err = writeApplianceValues(d, appliance); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func writeApplianceValues(d *schema.ResourceData, appliance *mcaasapi.Appliance) error {
	var err error

	createdDate, err := appliance.CreatedDate.MarshalText()
	if err != nil {
		return err
	}

	lastUpdateDate, err := appliance.LastUpdateDate.MarshalText()
	if err != nil {
		return err
	}

	if err = d.Set("created_date", string(createdDate)); err != nil {
		return err
	}

	if err = d.Set("last_update_date", string(lastUpdateDate)); err != nil {
		return err
	}

	if err = d.Set("name", appliance.Name); err != nil {
		return err
	}

	if err = d.Set("id", appliance.Id); err != nil {
		return err
	}

	return err
}
