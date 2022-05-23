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
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/utils"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

const (
	stateInitializing = "initializing"
	stateProvisioning = "infra-provisioning"
	stateCreating     = "creating"
	stateDeleting     = "deleting"
	stateReady        = "ready"
	stateDeleted      = "deleted"

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
		// TODO figure out if a cluster can be updated
		// Update:             clusterUpdate,
		DeleteContext: clusterDeleteContext,
		CustomizeDiff: nil,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(clusterAvailableTimeout),
			// Update: schema.DefaultTimeout(clusterAvailableTimeout),
			Delete: schema.DefaultTimeout(clusterDeleteTimeout),
		},
		Description: `The cluster resource facilitates the creation and
			deletion of a CaaS cluster.  Update is currently not supported.  There
			are four required inputs when creating a cluster - name, blueprint-id,
			site-id and space-id`,
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

	if err = d.Set("k8s_version", cluster.K8sVersion); err != nil {
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
