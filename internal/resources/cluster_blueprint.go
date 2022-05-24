// (C) Copyright 2020-2022 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

func ClusterBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:         nil,
		SchemaVersion:  0,
		StateUpgraders: nil,
		CreateContext:  clusterBlueprintCreateContext,
		ReadContext:    clusterBlueprintReadContext,
		// TODO figure out if and how a blueprint can be updated
		// Update:             clusterBlueprintUpdate,
		DeleteContext:      clusterBlueprintDeleteContext,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description:        `NOTE: this resource is currently not implemented`,
	}
}

func clusterBlueprintCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token: %s", err)
	}

	return nil
}

func clusterBlueprintReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token: %s", err)
	}

	return nil
}

func writeBlueprintResourceValues(d *schema.ResourceData, blueprint *mcaasapi.ClusterBlueprint) error {
	var err error

	createdDate, err := blueprint.CreatedDate.MarshalText()
	if err != nil {
		return err
	}

	lastUpdateDate, err := blueprint.LastUpdateDate.MarshalText()
	if err != nil {
		return err
	}

	if err = d.Set("created_date", string(createdDate)); err != nil {
		return err
	}

	if err = d.Set("last_update_date", string(lastUpdateDate)); err != nil {
		return err
	}

	if err = d.Set("name", blueprint.Name); err != nil {
		return err
	}

	if err = d.Set("k8s_version", blueprint.K8sVersion); err != nil {
		return err
	}

	if err = d.Set("cluster_provider", blueprint.ClusterProvider); err != nil {
		return err
	}

	machineSets := schemas.FlattenMachineSets(&blueprint.MachineSets)
	if err = d.Set("machine_sets", machineSets); err != nil {
		return err
	}

	if err = d.Set("default_storage_class", blueprint.DefaultStorageClass); err != nil {
		return err
	}

	return err
}

func clusterBlueprintDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token: %s", err)
	}

	return nil
}
