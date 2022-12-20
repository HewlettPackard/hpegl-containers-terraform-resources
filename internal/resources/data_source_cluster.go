// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Schema:             schemas.DataCluster(),
		ReadContext:        dataSourceClusterReadContext,
		SchemaVersion:      0,
		StateUpgraders:     nil,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `Cluster data source allows reading cluster data 
			based on name and space ID. Required inputs are name and space_id`,
	}
}

func dataSourceClusterReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	clusters, resp, err := c.CaasClient.ClustersApi.V1ClustersGet(clientCtx, field, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var cluster *mcaasapi.Cluster

	for b := range clusters.Items {
		if clusters.Items[b].Name == d.Get("name") {
			cluster = &clusters.Items[b]
			d.SetId(cluster.Id)
		}
	}

	if cluster == nil {
		return diag.Errorf("Cluster '%s' not found in space '%s'", d.Get("name"), spaceID)
	}

	if err = writeClusterResourceValues(d, cluster); err != nil {
		return diag.FromErr(err)
	}

	kubeconfig, _, err := c.CaasClient.KubeConfigApi.V1ClustersIdKubeconfigGet(clientCtx, cluster.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("kubeconfig", kubeconfig.Kubeconfig); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
