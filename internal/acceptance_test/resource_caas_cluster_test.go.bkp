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
	clusterName = "iac-acc-"
	blueprintID = "3f31daa9-9777-4c06-a4d0-e49215f5e48c"
	spaceID     = "f866c9bd-2d2c-4e60-aab0-64737df96273"
)

// nolint: gosec
func testCaasCluster() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return fmt.Sprintf(`
	provider hpegl {
		caas {
			api_url = "https://client.greenlake.hpe.com/api/caas/mcaas/v1"
		}
	}
	data "hpegl_caas_site" "blr" {
		name = "BLR"
		space_id = "%s"
	  }
	data "hpegl_caas_cluster_blueprint" "bp" {
		name = "demo"
		site_id = data.hpegl_caas_site.blr.id
	}
	resource hpegl_caas_cluster test {
		name         = "%s%d"
		blueprint_id = data.hpegl_caas_cluster_blueprint.bp.id
        site_id = data.hpegl_caas_site.blr.id
		space_id     = "%s"
	}`, spaceID, clusterName, r.Int63n(99999999), spaceID)
}

func TestCaasCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CaaS cluster creation in short mode.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(testCaasClusterDestroy("hpegl_caas_cluster.test")),
		Steps: []resource.TestStep{
			{
				Config: testCaasCluster(),
				Check:  resource.ComposeTestCheckFunc(checkCaasCluster("hpegl_caas_cluster.test")),
			},
		},
	})
}

func TestCaasPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testCaasCluster(),
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

func testCaasClusterDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["hpegl_caas_cluster.test"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_cluster.test")
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
