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

func DataSourceClusterProvider() *schema.Resource {
	return &schema.Resource{
		Schema:             schemas.ClusterProvider(),
		ReadContext:        dataSourceClusterProviderReadContext,
		SchemaVersion:      0,
		StateUpgraders:     nil,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `ClusterProvider data source allows reading Cluster Provider data 
			based on name and site ID. Required inputs are name and site ID`,
	}
}

func dataSourceClusterProviderReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	clusterProviders, resp, err := c.CaasClient.ClusterAdminApi.V1AppliancesIdClusterprovidersGet(clientCtx, applianceID)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var clusterProvider *mcaasapi.ClusterProvider

	for b := range clusterProviders.Items {
		if clusterProviders.Items[b].Name == d.Get("name") {
			clusterProvider = &clusterProviders.Items[b]
			d.SetId(clusterProvider.Id)
		}
	}

	if clusterProvider == nil {
		return diag.Errorf("Appliance '%s' not found in space '%s'", d.Get("name"), applianceID)
	}

	if err = writeClusterProviderValues(d, clusterProvider); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func writeClusterProviderValues(d *schema.ResourceData, clusterProvider *mcaasapi.ClusterProvider) error {
	var err error

	createdDate, err := clusterProvider.CreatedDate.MarshalText()
	if err != nil {
		return err
	}

	lastUpdateDate, err := clusterProvider.LastUpdateDate.MarshalText()
	if err != nil {
		return err
	}

	if err = d.Set("created_date", string(createdDate)); err != nil {
		return err
	}

	if err = d.Set("last_update_date", string(lastUpdateDate)); err != nil {
		return err
	}

	if err = d.Set("name", clusterProvider.Name); err != nil {
		return err
	}

	if err = d.Set("id", clusterProvider.Id); err != nil {
		return err
	}

	if err = d.Set("state", clusterProvider.State); err != nil {
		return err
	}

	if err = d.Set("health", clusterProvider.Health); err != nil {
		return err
	}

	storageClasses := schemas.FlattenStorageClasses(&clusterProvider.StorageClasses)
	if err = d.Set("storage_classes", storageClasses); err != nil {
		return err
	}

	minMasterSize := schemas.FlattenClusterProviderMinMasterSize(clusterProvider.MinMasterSize)
	if err = d.Set("min_master_size", minMasterSize); err != nil {
		return err
	}

	minWorkerSize := schemas.FlattenClusterProviderMinWorkerSize(clusterProvider.MinWorkerSize)
	if err = d.Set("min_worker_size", minWorkerSize); err != nil {
		return err
	}

	licenseInfo := schemas.FlattenLicenseInfo(clusterProvider.LicenseInfo)
	if err = d.Set("license_info", licenseInfo); err != nil {
		return err
	}

	availableCapacity := schemas.FlattenAvailableCapacity(clusterProvider.AvailableCapacity)
	if err = d.Set("available_capacity", availableCapacity); err != nil {
		return err
	}

	if err = d.Set("kubernetes_versions", clusterProvider.KubernetesVersions); err != nil {
		return err
	}
	return err
}
