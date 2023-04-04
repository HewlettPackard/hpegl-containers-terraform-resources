package resources

import (
	"context"
	"fmt"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

func MachineBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:         schemas.MachineBlueprintCreate(),
		SchemaVersion:  0,
		StateUpgraders: nil,
		CreateContext:  machineBlueprintCreateContext,
		ReadContext:    machineBlueprintReadContext,
		// TODO figure out if and how a blueprint can be updated
		// Update:             machineBlueprintUpdate,
		DeleteContext:      machineBlueprintDeleteContext,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description: `The machine blueprint resource facilitates the creation and
			deletion of a CaaS machine blueprint.  Update is currently not supported. The
			required inputs when creating a cluster blueprint are name,
			site-id, machine_provider, machine_roles, os_image, os_version, compute_type, size and storage_type`,
	}
}

func machineBlueprintCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	machineRoles := d.Get("machine_roles")
	machineRolesInt := machineRoles.([]interface{})

	var machineRolesStr []mcaasapi.MachineRolesType

	for _, val := range machineRolesInt {
		valStr := fmt.Sprint(val)
		machineRolesStr = append(machineRolesStr, mcaasapi.MachineRolesType(valStr))
	}

	workerType := mcaasapi.MachineWorkerType(d.Get("worker_type").(string))

	createMachineBlueprint := mcaasapi.MachineBlueprint{

		Name:                d.Get("name").(string),
		ApplianceID:         d.Get("site_id").(string),
		MachineRoles:        machineRolesStr,
		MachineProvider:     d.Get("machine_provider").(string),
		OsImage:             d.Get("os_image").(string),
		OsVersion:           d.Get("os_version").(string),
		ComputeInstanceType: d.Get("compute_type").(string),
		Size:                d.Get("size").(string),
		StorageInstanceType: d.Get("storage_type").(string),
		WorkerType:          &workerType,
	}

	machineBlueprint, resp, err := c.CaasClient.MachineBlueprintsApi.V1MachineblueprintsPost(clientCtx, createMachineBlueprint)
	if err != nil {
		errMessage := utils.GetErrorMessage(err, resp.StatusCode)
		diags = append(diags, diag.Errorf("Error in MachineBlueprintsPost: %s - %s", err, errMessage)...)

		return diags
	}
	defer resp.Body.Close()

	d.SetId(machineBlueprint.Id)

	return machineBlueprintReadContext(ctx, d, meta)
}

func machineBlueprintReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	applianceID := d.Get("site_id").(string)
	field := "applianceID eq " + applianceID
	machineBlueprint, resp, err := c.CaasClient.MachineBlueprintsApi.V1MachineblueprintsIdGet(clientCtx, id, field, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if err = writeMachineBlueprintResourceValues(d, &machineBlueprint); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func writeMachineBlueprintResourceValues(d *schema.ResourceData, machineBlueprint *mcaasapi.MachineBlueprint) error {
	var err error

	createdDate, err := machineBlueprint.CreatedDate.MarshalText()
	if err != nil {
		return err
	}

	lastUpdateDate, err := machineBlueprint.LastUpdateDate.MarshalText()
	if err != nil {
		return err
	}

	sizeDetail := schemas.FlattenSizeDetailMachineBlueprint(machineBlueprint.SizeDetail)

	if err = d.Set("created_date", string(createdDate)); err != nil {
		return err
	}

	if err = d.Set("last_update_date", string(lastUpdateDate)); err != nil {
		return err
	}

	if err = d.Set("name", machineBlueprint.Name); err != nil {
		return err
	}

	if err = d.Set("machine_provider", machineBlueprint.MachineProvider); err != nil {
		return err
	}

	if err = d.Set("machine_roles", machineBlueprint.MachineRoles); err != nil {
		return err
	}

	if err = d.Set("os_image", machineBlueprint.OsImage); err != nil {
		return err
	}

	if err = d.Set("os_version", machineBlueprint.OsVersion); err != nil {
		return err
	}

	if err = d.Set("size", machineBlueprint.Size); err != nil {
		return err
	}

	if err = d.Set("size_detail", sizeDetail); err != nil {
		return err
	}

	if err = d.Set("compute_type", machineBlueprint.ComputeInstanceType); err != nil {
		return err
	}

	if err = d.Set("storage_type", machineBlueprint.StorageInstanceType); err != nil {
		return err
	}

	if err = d.Set("site_id", machineBlueprint.ApplianceID); err != nil {
		return err
	}
	if err = d.Set("worker_type", machineBlueprint.WorkerType); err != nil {
		return err
	}

	return err
}

func machineBlueprintDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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

	resp, err := c.CaasClient.MachineBlueprintsApi.V1MachineblueprintsIdDelete(clientCtx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	d.SetId("")

	return diags

}
