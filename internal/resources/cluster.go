// (C) Copyright 2020-2023 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/resources/schemas"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/utils"
)

const (
	stateInitializing   = "initializing"
	stateProvisioning   = "infra-provisioning"
	stateDeProvisioning = "infra-deprovisioning"
	stateCreating       = "creating"
	stateDeleting       = "deleting"
	stateReady          = "ready"
	stateDeleted        = "deleted"
	stateUpdating       = "updating"
	stateUpgrading      = "upgrading"

	stateRetrying = "retrying" // placeholder state used to allow retrying after errors

	clusterAvailableTimeout = 60 * time.Minute
	clusterDeleteTimeout    = 60 * time.Minute
	pollingInterval         = 10 * time.Second

	// Number of retries if certain http response codes are returned by the client when polling
	// or if the cluster isn't present in the list of clusters (and we're not checking that the
	// cluster is deleted
	retryLimit = 3
)

// getTokenFunc type of function that is used to get a token, for use in polling loops
type getTokenFunc func() (string, error)

// nolint: funlen
func Cluster() *schema.Resource {
	return &schema.Resource{
		Schema:         schemas.Cluster(),
		SchemaVersion:  0,
		StateUpgraders: nil,
		CreateContext:  clusterCreateContext,
		ReadContext:    clusterReadContext,
		UpdateContext:  clusterUpdateContext,
		DeleteContext:  clusterDeleteContext,
		CustomizeDiff:  nil,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(clusterAvailableTimeout),
			Update: schema.DefaultTimeout(clusterAvailableTimeout),
			Delete: schema.DefaultTimeout(clusterDeleteTimeout),
		},
		Description: `The cluster resource facilitates the creation, updation and
			deletion of a CaaS cluster. There are four required inputs when 
			creating a cluster - name, blueprint_id, site_id and space_id. 
			worker_nodes is an optional input to scale nodes on cluster.
            OS Image update & Kubernetes version upgrade are also supported while updating the cluster.`,
	}
}

func clusterCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token in cluster-create: %s", err)
	}
	clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)

	var diags diag.Diagnostics

	spaceID := d.Get("space_id").(string)

	createCluster := mcaasapi.CreateCluster{
		Name:               d.Get("name").(string),
		ClusterBlueprintId: d.Get("blueprint_id").(string),
		ApplianceID:        d.Get("site_id").(string),
		SpaceID:            spaceID,
	}

	cluster, resp, err := c.CaasClient.ClustersApi.V1ClustersPost(clientCtx, createCluster)
	if err != nil {
		errMessage := utils.GetErrorMessage(err, resp.StatusCode)
		diags = append(diags, diag.Errorf("Error in ClustersPost: %s - %s", err, errMessage)...)

		return diags
	}
	defer resp.Body.Close()

	createStateConf := resource.StateChangeConf{
		Delay:      0,
		Pending:    []string{stateInitializing, stateProvisioning, stateCreating, stateRetrying},
		Target:     []string{stateReady},
		Timeout:    d.Timeout("create"),
		MinTimeout: pollingInterval,
		Refresh:    clusterRefresh(ctx, d, cluster.Id, spaceID, stateReady, meta),
	}

	_, err = createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Only set id to non-empty string if resource has been successfully created
	d.SetId(cluster.Id)

	// Set default master and worker nodes
	defaultFlattenMachineSets := schemas.FlattenMachineSets(&cluster.MachineSets)
	if err = d.Set("default_machine_sets", defaultFlattenMachineSets); err != nil {
		return diag.FromErr(err)
	}

	// Set default master and worker nodes details
	defaultFlattenMachineSetsDetail := schemas.FlattenMachineSetsDetail(&cluster.MachineSetsDetail)
	if err = d.Set("default_machine_sets_detail", defaultFlattenMachineSetsDetail); err != nil {
		return diag.FromErr(err)
	}

	//Add additional worker node pool after cluster creation
	workerNodes, workerNodePresent := d.GetOk("worker_nodes")
	newK8sVersionInterface, k8sVersionPresent := d.GetOk("kubernetesVersion")
	if workerNodePresent || k8sVersionPresent {
		workerNodesList := workerNodes.([]interface{})
		machineSets := []mcaasapi.MachineSet{}

		for _, workerNode := range workerNodesList {
			machineSets = append(machineSets, getWorkerNodeDetails(d, workerNode.(map[string]interface{})))
		}

		defaultMachineSets := cluster.MachineSets
		defaultWorkersName, err := GetDefaultWorkersName(d)
		if err != nil {
			return diag.FromErr(err)
		}
		//Remove default worker node if its declared in worker nodes
		for _, defaultWorkerName := range defaultWorkersName {
			if utils.WorkerPresentInMachineSets(machineSets, defaultWorkerName) {
				defaultMachineSets = utils.RemoveWorkerFromMachineSets(defaultMachineSets, defaultWorkerName)
			}
		}

		machineSets = append(defaultMachineSets, machineSets...)
		temp, err := json.Marshal(machineSets)
		if err != nil {
			return diag.Errorf("Error in parsing machinesets response %s", err)
		}
		var finalMachineSets []mcaasapi.UpdateClusterMachineSet
		_ = json.Unmarshal(temp, &finalMachineSets)

		//Check if kubernetesVersion update is present
		newK8sVersion := ""
		if k8sVersionPresent {
			newK8sVersion = fmt.Sprintf("%v", newK8sVersionInterface)
		}

		updateCluster := mcaasapi.UpdateCluster{
			MachineSets:       finalMachineSets,
			KubernetesVersion: newK8sVersion,
		}

		clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)
		cluster, resp, err := c.CaasClient.ClustersApi.V1ClustersIdPut(clientCtx, updateCluster, cluster.Id)
		if err != nil {
			errMessage := utils.GetErrorMessage(err, resp.StatusCode)
			diags = append(diags, diag.Errorf("Error in V1ClustersIdPut: %s - %s", err, errMessage)...)
			return diags
		}
		defer resp.Body.Close()

		createStateConf := resource.StateChangeConf{
			Delay:      0,
			Pending:    []string{stateProvisioning, stateCreating, stateRetrying, stateUpdating, stateDeProvisioning, stateUpgrading},
			Target:     []string{stateReady},
			Timeout:    d.Timeout("create"),
			MinTimeout: pollingInterval,
			Refresh:    clusterRefresh(ctx, d, cluster.Id, spaceID, stateReady, meta),
		}

		_, err = createStateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// TODO Should we be passing clientCtx here?
	return clusterReadContext(ctx, d, meta)
}

func clusterRefresh(ctx context.Context, d *schema.ResourceData,
	id, spaceID, expectedState string,
	meta interface{},
) resource.StateRefreshFunc {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return func() (interface{}, string, error) { return nil, "", err }
	}

	// Create getTokenFunc for execution in a closure that increments retry counters
	gtf := createGetTokenFunc(ctx, c, id, spaceID, expectedState, meta)

	return func() (result interface{}, state string, err error) {
		state, err = gtf()

		return d.Get("name"), state, err
	}
}

func clusterReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	spaceID := d.Get("space_id").(string)
	field := "spaceID eq " + spaceID
	cluster, resp, err := c.CaasClient.ClustersApi.V1ClustersIdGet(clientCtx, id, field)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if err = writeClusterResourceValues(d, &cluster); err != nil {
		return diag.FromErr(err)
	}

	kubeconfig, _, err := c.CaasClient.KubeConfigApi.V1ClustersIdKubeconfigGet(clientCtx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("kubeconfig", kubeconfig.Kubeconfig); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// nolint: cyclop
func writeClusterResourceValues(d *schema.ResourceData, cluster *mcaasapi.Cluster) error {
	var err error
	if err = d.Set("state", cluster.State); err != nil {
		return err
	}

	if err = d.Set("health", cluster.Health); err != nil {
		return err
	}

	createdDate, err := cluster.CreatedDate.MarshalText()
	if err != nil {
		return err
	}

	lastUpdateDate, err := cluster.LastUpdateDate.MarshalText()
	if err != nil {
		return err
	}

	if err = d.Set("created_date", string(createdDate)); err != nil {
		return err
	}

	if err = d.Set("last_update_date", string(lastUpdateDate)); err != nil {
		return err
	}

	if err = d.Set("name", cluster.Name); err != nil {
		return err
	}

	if err = d.Set("blueprint_id", cluster.ClusterBlueprintId); err != nil {
		return err
	}

	if err = d.Set("kubernetes_version", cluster.KubernetesVersion); err != nil {
		return err
	}

	if err = d.Set("cluster_provider", cluster.ClusterProvider); err != nil {
		return err
	}

	machineSets := schemas.FlattenMachineSets(&cluster.MachineSets)
	if err = d.Set("machine_sets", machineSets); err != nil {
		return err
	}

	machineSetsDetail := schemas.FlattenMachineSetsDetail(&cluster.MachineSetsDetail)
	if err = d.Set("machine_sets_detail", machineSetsDetail); err != nil {
		return err
	}

	if err = d.Set("api_endpoint", cluster.ApiEndpoint); err != nil {
		return err
	}

	serviceEndpoints := schemas.FlattenServiceEndpoints(&cluster.ServiceEndpoints)
	if err = d.Set("service_endpoints", serviceEndpoints); err != nil {
		return err
	}

	if err = d.Set("site_id", cluster.ApplianceID); err != nil {
		return err
	}

	if err = d.Set("appliance_name", cluster.ApplianceName); err != nil {
		return err
	}

	if err = d.Set("space_id", cluster.SpaceID); err != nil {
		return err
	}

	if err = d.Set("default_storage_class", cluster.DefaultStorageClass); err != nil {
		return err
	}

	if err = d.Set("default_storage_class_description", cluster.DefaultStorageClassDescription); err != nil {
		return err
	}

	return err
}

func clusterDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	spaceID := d.Get("space_id").(string)

	_, resp, err := c.CaasClient.ClustersApi.V1ClustersIdDelete(clientCtx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	deleteStateConf := resource.StateChangeConf{
		Delay:      pollingInterval,
		Pending:    []string{stateDeleting, stateRetrying},
		Target:     []string{stateDeleted},
		Timeout:    d.Timeout("delete"),
		MinTimeout: pollingInterval,
		Refresh:    clusterRefresh(ctx, d, id, spaceID, stateDeleted, meta),
	}

	_, err = deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Only set id to "" if delete has been successful, this means that terraform will delete the resource entry
	// This also means that the destroy can be reattempted by terraform if there was an error
	d.SetId("")

	return diags
}

// createGetTokenFunc is a closure that returns a getTokenFunc
// The closure sets counters that are incremented on each execution of getTokenFunc
// nolint cyclop
func createGetTokenFunc(
	ctx context.Context,
	c *client.Client,
	id, spaceID, expectedState string,
	meta interface{},
) getTokenFunc {
	// We set these counters in the closure
	noEntryInListRetryCount := 0
	errRetryCount := 0

	return func() (string, error) {
		var cluster *mcaasapi.Cluster
		// Get token - we run this on every loop iteration in case the token is about
		// to expire
		token, err := auth.GetToken(ctx, meta)
		if err != nil {
			return "", err
		}
		clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)
		field := "spaceID eq " + spaceID
		clusters, resp, err := c.CaasClient.ClustersApi.V1ClustersGet(clientCtx, field)
		if err != nil {
			if resp != nil {
				// Check err response code to see if we need to retry
				switch resp.StatusCode {
				// TODO we've added this since at the moment CaaS returns 500 on IAM timeout, they will return 429
				case http.StatusInternalServerError:
					errRetryCount++
					if errRetryCount < retryLimit {
						return stateRetrying, nil
					}

					fallthrough

				case http.StatusGatewayTimeout:
					errRetryCount++
					if errRetryCount < retryLimit {
						return stateRetrying, nil
					}

					fallthrough

				default:
					return "", err
				}
			}

			if isErrRetryable(err) {
				errRetryCount++
				if errRetryCount < retryLimit {
					return stateRetrying, nil
				}
			}

			// Error not retryable, exit
			return "", errors.New("error in getting cluster list: " + err.Error())
		}
		// Reset error counter
		errRetryCount = 0
		defer resp.Body.Close()

		for i := range clusters.Items {
			if clusters.Items[i].Id == id {
				cluster = &clusters.Items[i]
			}
		}

		// cluster doesn't exist, check if we expect it to be deleted
		if cluster == nil {
			switch expectedState {
			case stateDeleted:
				return stateDeleted, nil

			default:
				noEntryInListRetryCount++
				if noEntryInListRetryCount > retryLimit {
					return "", errors.New("failed to find cluster in list")
				}

				return stateRetrying, nil
			}
		}
		// Reset noEntryInListRetryCount
		noEntryInListRetryCount = 0

		return cluster.State, nil
	}
}

// isErrRetryable checks if an error is retryable, currently limited to net Timeout errors
func isErrRetryable(err error) bool {
	var t net.Error
	if errors.As(err, &t) && t.Timeout() {
		return true
	}

	return false
}

func clusterUpdateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := auth.GetToken(ctx, meta)
	if err != nil {
		return diag.Errorf("Error in getting token in cluster-create: %s", err)
	}

	clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)
	var diags diag.Diagnostics
	newK8sVersionInterface, k8sVersionPresent := d.GetOk("kubernetes_version")

	if d.HasChange("worker_nodes") || k8sVersionPresent {
		machineSets := []mcaasapi.MachineSet{}

		workerNodes := d.Get("worker_nodes").([]interface{})
		for _, workerNode := range workerNodes {
			machineSets = append(machineSets, getWorkerNodeDetails(d, workerNode.(map[string]interface{})))
		}

		defaultMachineSetsInterface := d.Get("default_machine_sets").([]interface{})
		defaultMachineSets := []mcaasapi.MachineSet{}

		for _, dms := range defaultMachineSetsInterface {
			defaultMachineSet := getDefaultMachineSet(d, dms.(map[string]interface{}))
			defaultMachineSets = append(defaultMachineSets, defaultMachineSet)
		}
		defaultWorkersName, err := GetDefaultWorkersName(d)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, defaultWorkerName := range defaultWorkersName {
			if utils.WorkerPresentInMachineSets(machineSets, defaultWorkerName) {
				defaultMachineSets = utils.RemoveWorkerFromMachineSets(defaultMachineSets, defaultWorkerName)
			}
		}
		machineSets = append(machineSets, defaultMachineSets...)
		temp, err := json.Marshal(machineSets)
		if err != nil {
			return diag.Errorf("Error in parsing machinesets response %s", err)
		}
		var finalMachineSets []mcaasapi.UpdateClusterMachineSet
		_ = json.Unmarshal(temp, &finalMachineSets)
		//Check if kubernetesVersion update is present
		newK8sVersion := ""
		if k8sVersionPresent {
			newK8sVersion = fmt.Sprintf("%v", newK8sVersionInterface)
		}

		updateCluster := mcaasapi.UpdateCluster{
			MachineSets:       finalMachineSets,
			KubernetesVersion: newK8sVersion,
		}
		clusterID := d.Id()
		cluster, resp, err := c.CaasClient.ClustersApi.V1ClustersIdPut(clientCtx, updateCluster, clusterID)
		if err != nil {
			errMessage := utils.GetErrorMessage(err, resp.StatusCode)
			diags = append(diags, diag.Errorf("Error in V1ClustersIdPut: %s - %s", err, errMessage)...)
			return diags
		}
		defer resp.Body.Close()

		spaceID := d.Get("space_id").(string)
		createStateConf := resource.StateChangeConf{
			Delay:      0,
			Pending:    []string{stateProvisioning, stateCreating, stateRetrying, stateUpdating, stateDeProvisioning, stateUpgrading},
			Target:     []string{stateReady},
			Timeout:    d.Timeout("create"),
			MinTimeout: pollingInterval,
			Refresh:    clusterRefresh(ctx, d, cluster.Id, spaceID, stateReady, meta),
		}

		_, err = createStateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return clusterReadContext(ctx, d, meta)
}

func getDefaultMachineSet(d *schema.ResourceData, defaultMachineSet map[string]interface{}) mcaasapi.MachineSet {
	//Use the updated os Image version in the update request body so that at the time of scale up/down, the update request body has the latest os image version.
	osVersion := ""
	osImage := ""
	dwns, _ := GetDefaultWorkersName(d)
	var found bool
	for _, dwn := range dwns {
		if defaultMachineSet["name"].(string) == dwn {
			found = true
			break
		}
	}
	if found {
		for _, dwn := range dwns {
			if defaultMachineSet["name"].(string) == dwn {
				machinesets := d.Get("machine_sets").([]interface{})
				for _, machinesetInt := range machinesets {
					machineset := machinesetInt.(map[string]interface{})
					if dwn == machineset["name"] {
						osVersion = fmt.Sprintf("%v", machineset["os_version"])
						osImage = fmt.Sprintf("%v", machineset["os_image"])
					}
				}
			}
		}

	} else {
		osVersion = defaultMachineSet["os_version"].(string)
		osImage = defaultMachineSet["os_image"].(string)

	}
	wn := mcaasapi.MachineSet{
		MachineBlueprintId: defaultMachineSet["machine_blueprint_id"].(string),
		Count:              int32(defaultMachineSet["count"].(float64)),
		Name:               defaultMachineSet["name"].(string),
		OsImage:            osImage,
		OsVersion:          osVersion,
	}
	return wn
}
func getDefaultMachineSetDetail(defaultMachineSetDetail map[string]interface{}) mcaasapi.MachineSetDetail {
	mr := defaultMachineSetDetail["machine_roles"].([]interface{})
	MachineRoles := make([]mcaasapi.MachineRolesType, 0, len(mr))
	for _, v := range mr {
		MachineRoles = append(MachineRoles, mcaasapi.MachineRolesType(v.(string)))
	}
	macProvider := defaultMachineSetDetail["machine_provider"]
	machineProvider := mcaasapi.MachineProviderName(macProvider.(string))
	wnd := mcaasapi.MachineSetDetail{
		Name:                defaultMachineSetDetail["name"].(string),
		OsImage:             defaultMachineSetDetail["os_image"].(string),
		OsVersion:           defaultMachineSetDetail["os_version"].(string),
		Count:               int32(defaultMachineSetDetail["count"].(float64)),
		MachineRoles:        MachineRoles,
		MachineProvider:     &machineProvider,
		Size:                defaultMachineSetDetail["size"].(string),
		ComputeInstanceType: defaultMachineSetDetail["compute_type"].(string),
		StorageInstanceType: defaultMachineSetDetail["storage_type"].(string),
	}
	return wnd
}

func GetDefaultWorkersName(d *schema.ResourceData) ([]string, error) {

	defaultMachineSetsDetailInterface := d.Get("default_machine_sets_detail").([]interface{})
	defaultMachineSetsDetail := []mcaasapi.MachineSetDetail{}

	for _, dmsd := range defaultMachineSetsDetailInterface {
		defaultMachineSetDetail := getDefaultMachineSetDetail(dmsd.(map[string]interface{}))
		defaultMachineSetsDetail = append(defaultMachineSetsDetail, defaultMachineSetDetail)
	}

	var workerNames []string

	for _, msd := range defaultMachineSetsDetail {
		for _, role := range msd.MachineRoles {
			if role == "worker" {
				workerNames = append(workerNames, msd.Name)
			}
		}
	}
	if len(workerNames) == 0 {
		return nil, fmt.Errorf("Worker node not present in the cluster")
	}
	return workerNames, nil
}
