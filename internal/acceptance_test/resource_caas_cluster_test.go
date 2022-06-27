// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP.

package acceptancetest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/internal/utils"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

const (
	clusterPrefix  = "test"
	apiURL         = "https://mcaas.us1.greenlake-hpe.com/mcaas"
	siteName       = "Austin"
	testWorkerNode = "testworkernode"
)

// nolint: gosec
func testCaasCluster(clusterName string) string {
	return fmt.Sprintf(`
	provider hpegl {
		caas {
			api_url = "%s"
		}
	}
	variable "HPEGL_SPACE" {
  		type = string
	}
		data "hpegl_caas_site" "site" {
			name = "%s"
			space_id = var.HPEGL_SPACE
		}
		data "hpegl_caas_cluster_blueprint" "bp" {
			name = "demo"
			site_id = data.hpegl_caas_site.site.id
		}
	resource hpegl_caas_cluster testcluster {
		name         = "%v"
		blueprint_id = data.hpegl_caas_cluster_blueprint.bp.id
        site_id = data.hpegl_caas_site.site.id
		space_id     = var.HPEGL_SPACE
	}`, apiURL, siteName, clusterName)
}

// nolint: gosec
func testCaasClusterUpdate(clusterName string) string {
	return fmt.Sprintf(`
	provider hpegl {
		caas {
			api_url = "%s"
		}
	}
	variable "HPEGL_SPACE" {
  		type = string
	}
		data "hpegl_caas_site" "site" {
			name = "%s"
			space_id = var.HPEGL_SPACE
		}
		data "hpegl_caas_cluster_blueprint" "bp" {
			name = "demo"
			site_id = data.hpegl_caas_site.site.id
		}
		data "hpegl_caas_machine_blueprint" "mbworker" {
			name = "standard-worker"
		site_id = data.hpegl_caas_site.site.id
	  }
	resource hpegl_caas_cluster testcluster {
		name         = "%v"
		blueprint_id = data.hpegl_caas_cluster_blueprint.bp.id
        site_id = data.hpegl_caas_site.site.id
		space_id     = var.HPEGL_SPACE
		
		worker_nodes {
			name = "%s"
			machine_blueprint_id = data.hpegl_caas_machine_blueprint.mbworker.id
			count = "1"
		  }
	}`, apiURL, siteName, clusterName, testWorkerNode)
}

func TestCaasCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CaaS cluster creation in short mode.")
	}

	clusterName := fmt.Sprintf("%s-%s", clusterPrefix, randomHex(5))

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              resource.ComposeTestCheckFunc(testCaasClusterDestroy("hpegl_caas_cluster.testcluster")),
		Steps: []resource.TestStep{
			{
				Config: testCaasCluster(clusterName),
				Check:  resource.ComposeTestCheckFunc(checkCaasCluster("hpegl_caas_cluster.testcluster")),
			},
			{
				Config: testCaasClusterUpdate(clusterName),
				Check:  resource.ComposeTestCheckFunc(checkCaasCluster("hpegl_caas_cluster.testcluster"), checkCaasClusterUpdate("hpegl_caas_cluster.testcluster")),
			},
		},
	})
}

func TestCaasPlan(t *testing.T) {
	clusterName := fmt.Sprintf("%s-%s", clusterPrefix, randomHex(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testCaasCluster(clusterName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkCaasCluster(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found: %s", name)
		}

		state := rs.Primary.Attributes["state"]
		if state != "ready" {
			return fmt.Errorf("Cluster not ready")
		}

		return nil
	}
}

func checkCaasClusterUpdate(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found: %s", name)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		id := rs.Primary.Attributes["id"]

		p, err := client.GetClientFromMetaMap(testAccProvider.Meta())
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		token, err := auth.GetToken(ctx, testAccProvider.Meta())
		if err != nil {
			return fmt.Errorf("Failed getting a token: %w", err)
		}
		clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)

		cluster, _, err := p.CaasClient.ClusterAdminApi.V1ClustersIdGet(clientCtx, id, spaceID)
		if err != nil {
			return fmt.Errorf("Error in getting cluster list %w", err)
		}

		if len(cluster.MachineSets) != 3 {
			return fmt.Errorf("Incorrect worker and master nodes, expected 3 found %v", len(cluster.MachineSets))
		}

		if !utils.WorkerPresentInMachineSets(cluster.MachineSets, testWorkerNode) {
			return fmt.Errorf("Worker node pool %v not present in cluster %v", testWorkerNode, cluster.Name)
		}

		return nil
	}
}

func testCaasClusterDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["hpegl_caas_cluster.testcluster"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_cluster.testcluster")
		}

		spaceID := rs.Primary.Attributes["space_id"]

		p, err := client.GetClientFromMetaMap(testAccProvider.Meta())
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		token, err := auth.GetToken(ctx, testAccProvider.Meta())
		if err != nil {
			return fmt.Errorf("Failed getting a token: %w", err)
		}
		clientCtx := context.WithValue(ctx, mcaasapi.ContextAccessToken, token)

		var cluster *mcaasapi.Cluster
		clusters, _, err := p.CaasClient.ClusterAdminApi.V1ClustersGet(clientCtx, spaceID)
		if err != nil {
			return fmt.Errorf("Error in getting cluster list %w", err)
		}

		for i := range clusters.Items {
			if clusters.Items[i].Id == rs.Primary.ID {
				cluster = &clusters.Items[i]
			}
		}

		if cluster != nil {
			return fmt.Errorf("Cluster still exists")
		}

		return nil
	}
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
