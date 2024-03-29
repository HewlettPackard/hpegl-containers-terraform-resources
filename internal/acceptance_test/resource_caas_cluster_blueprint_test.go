// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP.

package acceptancetest

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/HewlettPackard/hpegl-containers-go-sdk/pkg/mcaasapi"

	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/auth"
	"github.com/HewlettPackard/hpegl-containers-terraform-resources/pkg/client"
)

const (
	// Fill in these values based on the environment being used for acceptance testing
	name                = "test-cluster-bp"
	defaultStorageClass = "gl-sbc-glhc-nimblestor"
	clusterProvider     = "ecp"
	cpCount             = "1"
	workerName          = "worker1"
	//workerCount         = "1"
	workerMinSize = "1"
	workerMaxSize = "1"
	//kubernetesVersion   = "1.24.10-hpe1"
	apiURLCbp   = "https://mcaas.intg.hpedevops.net/mcaas"
	siteNameCBp = "FTC"
	//apiURLCbp           = "https://mcaas.us1.greenlake-hpe.com/mcaas"
)

// nolint: gosec
func testCaasClusterBlueprint() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

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
    data "hpegl_caas_machine_blueprint" "mbcontrolplane" {
  		name = "large-master"
  		site_id = data.hpegl_caas_site.site.id
	}
	data "hpegl_caas_machine_blueprint" "mbworker" {
  		name = "xlarge-worker"
  		site_id = data.hpegl_caas_site.site.id
	}
    data "hpegl_caas_cluster_provider" "clusterprovider" {
        name = "ecp"
        site_id = data.hpegl_caas_site.site.id
    }
	resource hpegl_caas_cluster_blueprint testcb {
		name         = "%s%d"
		kubernetes_version  = data.hpegl_caas_cluster_provider.clusterprovider.kubernetes_versions[0]
  		default_storage_class = "%s"
  		site_id = data.hpegl_caas_site.site.id
  		cluster_provider = "%s"
		control_plane_count = "%s"
  		worker_nodes {
			name = "%s"
      		machine_blueprint_id = data.hpegl_caas_machine_blueprint.mbworker.id
			min_size = "%s"
			max_size = "%s"
    	}
	}`, apiURLCbp, siteNameCBp, name, r.Int63n(99999999), defaultStorageClass, clusterProvider, cpCount, workerName, workerMinSize, workerMaxSize)
}

func TestCaasClusterBlueprintCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              resource.ComposeTestCheckFunc(testCaasClusterBlueprintDestroy("hpegl_caas_cluster_blueprint.testcb")),
		Steps: []resource.TestStep{
			{
				Config: testCaasClusterBlueprint(),
				Check:  resource.ComposeTestCheckFunc(checkCaasClusterBlueprint("hpegl_caas_cluster_blueprint.testcb")),
			},
		},
	})
}

func TestCaasClusterBlueprintPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testCaasClusterBlueprint(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkCaasClusterBlueprint(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("ClusterBlueprint not found: %s", name)
		}
		return nil
	}
}

func testCaasClusterBlueprintDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["hpegl_caas_cluster_blueprint.testcb"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_cluster_blueprint.testcb")
		}

		siteID := rs.Primary.Attributes["site_id"]

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

		var clusterBlueprint *mcaasapi.ClusterBlueprint
		field := "applianceID eq " + siteID
		clusterBlueprints, _, err := p.CaasClient.ClusterBlueprintsApi.V1ClusterblueprintsGet(clientCtx, field)
		if err != nil {
			return fmt.Errorf("Error in getting cluster blueprint list %w", err)
		}

		for i := range clusterBlueprints.Items {
			if clusterBlueprints.Items[i].Id == rs.Primary.ID {
				clusterBlueprint = &clusterBlueprints.Items[i]
			}
		}

		if clusterBlueprint != nil {
			return fmt.Errorf("ClusterBlueprint still exists")
		}

		return nil
	}
}
