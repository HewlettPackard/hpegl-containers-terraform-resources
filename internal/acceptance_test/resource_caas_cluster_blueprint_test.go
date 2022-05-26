// (C) Copyright 2020-2021 Hewlett Packard Enterprise Development LP.

package acceptancetest

import (
	"context"
	"fmt"
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
	name                        = "tf-bp-test"
	k8s_version                 = "v1.20.11.hpe-2"
	default_storage_class       = ""
	spaceId                     = ""
	cluster_provider            = "ecp"
	cp_machine_blueprint_id     = ""
	cp_count                    = "1"
	worker_name                 = "worker1"
	worker_machine_blueprint_id = ""
	worker_count                = "1"
	site_id                     = ""
)

// nolint: gosec
func testCaasClusterBlueprint() string {

	return fmt.Sprintf(`
	provider hpegl {
		caas {
			api_url = "https://client.greenlake.hpe.com/api/caas/mcaas"
		}
	}
	data "hpegl_caas_site" "blr" {
		name = "BLR"
		space_id = "%s"
	  }
	resource hpegl_caas_cluster test {
		name         = "%s"
  		k8s_version  = "%s"
  		default_storage_class = "%s"
  		site_id = data.hpegl_caas_site.blr.id
  		cluster_provider = "%s"
		control_plane_nodes = {
    		machine_blueprint_id = "%s"
			count = "%s"
  		}
  		worker_nodes {
			name = "%s"
      		machine_blueprint_id = "%s"
      		count = "%s"
    	}
	}`, spaceId, name, k8s_version, default_storage_class, cluster_provider, cp_machine_blueprint_id, cp_count, worker_name, worker_machine_blueprint_id, worker_count)
}

func TestCaasClusterBlueprintCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(testCaasClusterBlueprintDestroy("hpegl_caas_cluster_blueprint.testbp")),
		Steps: []resource.TestStep{
			{
				Config: testCaasClusterBlueprint(),
				Check:  resource.ComposeTestCheckFunc(checkCaasClusterBlueprint("hpegl_caas_cluster_blueprint.testbp")),
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
		rs, ok := s.RootModule().Resources["hpegl_caas_cluster_blueprint.testbp"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_cluster_blueprint.testbp")
		}

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
		clusterBlueprints, _, err := p.CaasClient.ClusterAdminApi.V1ClusterblueprintsGet(clientCtx, site_id)
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