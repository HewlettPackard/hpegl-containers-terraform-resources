// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"
	"errors"
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
	stateInitializing = "initializing"
	stateProvisioning = "infra-provisioning"
	stateCreating     = "creating"
	stateDeleting     = "deleting"
	stateReady        = "ready"
	stateDeleted      = "deleted"
	stateUpdating     = "updating"

	stateRetrying = "retrying" // placeholder state used to allow retrying after errors

	clusterAvailableTimeout = 60 * time.Minute
	clusterDeleteTimeout    = 60 * time.Minute
	pollingInterval         = 10 * time.Second

	// Number of retries if certain http response codes are returned by the client when polling
	// or if the cluster isn't present in the list of clusters (and we're not checking that the
	// cluster is deleted
	retryLimit = 3

	//Default worker Node Pool Name
	defaultWorkerName = "worker"
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
			worker_nodes is an optional input to scale nodes on cluster.`,
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

	cluster, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersPost(clientCtx, createCluster)
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
		Timeout:    clusterAvailableTimeout,
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

	//Add additional worker node pool after cluster creation
	workerNodes, workerNodePresent := d.GetOk("worker_nodes")
	if workerNodePresent {
		workerNodesList := workerNodes.([]interface{})
		machineSets := []mcaasapi.MachineSet{}

		for _, workerNode := range workerNodesList {
			machineSets = append(machineSets, getWorkerNodeDetails(workerNode.(map[string]interface{})))
		}

		defaultMachineSets := cluster.MachineSets
		//Remove default worker node if its declared in worker nodes
		if utils.WorkerPresentInMachineSets(machineSets, defaultWorkerName) {
			defaultMachineSets = utils.RemoveWorkerFromMachineSets(cluster.MachineSets, defaultWorkerName)
		}

		machineSets = append(defaultMachineSets, machineSets...)
		updateCluster := mcaasapi.UpdateCluster{
			MachineSets: machineSets,
		}

		clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)
		cluster, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersIdPut(clientCtx, updateCluster, cluster.Id)
		if err != nil {
			errMessage := utils.GetErrorMessage(err, resp.StatusCode)
			diags = append(diags, diag.Errorf("Error in V1ClustersIdPut: %s - %s", err, errMessage)...)
			return diags
		}
		defer resp.Body.Close()

		createStateConf := resource.StateChangeConf{
			Delay:      0,
			Pending:    []string{stateProvisioning, stateCreating, stateRetrying, stateUpdating},
			Target:     []string{stateReady},
			Timeout:    clusterAvailableTimeout,
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

	cluster, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersIdGet(clientCtx, id, spaceID)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if err = writeClusterResourceValues(d, &cluster); err != nil {
		return diag.FromErr(err)
	}

	kubeconfig, _, err := c.CaasClient.ClusterAdminApi.V1ClustersIdKubeconfigGet(clientCtx, id)
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

	_, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersIdDelete(clientCtx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	deleteStateConf := resource.StateChangeConf{
		Delay:      pollingInterval,
		Pending:    []string{stateDeleting, stateRetrying},
		Target:     []string{stateDeleted},
		Timeout:    clusterDeleteTimeout,
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

		clusters, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersGet(clientCtx, spaceID)
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

	if d.HasChange("worker_nodes") {
		machineSets := []mcaasapi.MachineSet{}

		workerNodes := d.Get("worker_nodes").([]interface{})
		for _, workerNode := range workerNodes {
			machineSets = append(machineSets, getWorkerNodeDetails(workerNode.(map[string]interface{})))
		}

		defaultMachineSetsInterface := d.Get("default_machine_sets").([]interface{})
		defaultMachineSets := []mcaasapi.MachineSet{}

		for _, dms := range defaultMachineSetsInterface {
			defaultMachineSet := getDefaultMachineSet(dms.(map[string]interface{}))
			defaultMachineSets = append(defaultMachineSets, defaultMachineSet)
		}

		if utils.WorkerPresentInMachineSets(machineSets, defaultWorkerName) {
			defaultMachineSets = utils.RemoveWorkerFromMachineSets(defaultMachineSets, defaultWorkerName)
		}

		machineSets = append(machineSets, defaultMachineSets...)

		updateCluster := mcaasapi.UpdateCluster{
			MachineSets: machineSets,
		}
		clusterID := d.Id()
		cluster, resp, err := c.CaasClient.ClusterAdminApi.V1ClustersIdPut(clientCtx, updateCluster, clusterID)
		if err != nil {
			errMessage := utils.GetErrorMessage(err, resp.StatusCode)
			diags = append(diags, diag.Errorf("Error in V1ClustersIdPut: %s - %s", err, errMessage)...)
			return diags
		}
		defer resp.Body.Close()

		spaceID := d.Get("space_id").(string)
		createStateConf := resource.StateChangeConf{
			Delay:      0,
			Pending:    []string{stateProvisioning, stateCreating, stateRetrying, stateUpdating},
			Target:     []string{stateReady},
			Timeout:    clusterAvailableTimeout,
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

func getDefaultMachineSet(defaultMachineSet map[string]interface{}) mcaasapi.MachineSet {
	wn := mcaasapi.MachineSet{
		MachineBlueprintId: defaultMachineSet["machine_blueprint_id"].(string),
		Count:              defaultMachineSet["count"].(float64),
		Name:               defaultMachineSet["name"].(string),
		OsImage:            defaultMachineSet["os_image"].(string),
		OsVersion:          defaultMachineSet["os_version"].(string),
	}
	return wn
}
