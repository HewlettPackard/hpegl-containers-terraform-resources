// (C) Copyright 2020-2023 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"
	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ClusterBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:         schemas.ClusterBlueprintCreate(),
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
		Description: `The cluster blueprint resource facilitates the creation and
			deletion of a CaaS cluster blueprint.  Update is currently not supported. The
			required inputs when creating a cluster blueprint are name, kubernetes_version,
			site-id, cluster_provider, control_plane, worker_nodes and default_storage_class`,
	}
}

func clusterBlueprintCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token in cluster-blueprint-create: %s", err)
	}
	clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)

	var diags diag.Diagnostics
	var machineSetsList []mcaasapi.MachineSet

	workerNodesList := d.Get("worker_nodes").([]interface{})
	for _, workerNode := range workerNodesList {
		worker, ok := workerNode.(map[string]interface{})
		if ok {
			workerNodeDetails := getWorkerNodeDetails(worker)
			machineSetsList = append(machineSetsList, workerNodeDetails)
		}
	}

	createClusterBlueprint := mcaasapi.ClusterBlueprint{
		Name:                d.Get("name").(string),
		KubernetesVersion:   d.Get("kubernetes_version").(string),
		DefaultStorageClass: d.Get("default_storage_class").(string),
		ApplianceID:         d.Get("site_id").(string),
		ClusterProvider:     d.Get("cluster_provider").(string),
		ControlPlaneCount:   int32(d.Get("control_plane_count").(float64)),
		MachineSets:         machineSetsList,
	}

	clusterBlueprint, resp, err := c.CaasClient.ClusterBlueprintsApi.V1ClusterblueprintsPost(clientCtx, createClusterBlueprint)
	if err != nil {
		errMessage := utils.GetErrorMessage(err, resp.StatusCode)
		diags = append(diags, diag.Errorf("Error in ClustersBlueprintPost: %s - %s", err, errMessage)...)

		return diags
	}
	defer resp.Body.Close()

	d.SetId(clusterBlueprint.Id)

	return clusterBlueprintReadContext(ctx, d, meta)
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

	if err = d.Set("kubernetes_version", blueprint.KubernetesVersion); err != nil {
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
	id := d.Id()

	resp, err := c.CaasClient.ClusterBlueprintsApi.V1ClusterblueprintsIdDelete(clientCtx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	d.SetId("")

	return diags
}

func getWorkerNodeDetails(workerNode map[string]interface{}) mcaasapi.MachineSet {

	wn := mcaasapi.MachineSet{
		MachineBlueprintId: workerNode["machine_blueprint_id"].(string),
		MinSize:            int32(workerNode["min_size"].(float64)),
		MaxSize:            int32(workerNode["max_size"].(float64)),
		Name:               workerNode["name"].(string),
	}
	return wn
}
