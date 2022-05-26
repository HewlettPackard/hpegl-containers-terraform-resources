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
	nameMbp                     = "mbp-test"
	siteIDMbp                   = ""
	machineProvider             = "vmaas"
	osImage                     = "sles-custom"
	osVersion                   = ""
	computeType                 = ""
	size                        = ""
	storageType                 = ""
)

var machineRoles = []string{"controlplane"}

// nolint: gosec
func testCaasMachineBlueprint() string {

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
	resource hpegl_caas_machine_blueprint test {
		name         = "%s"
  		site_id = data.hpegl_caas_site.blr.id
  		machine_roles = ["%s"]
		machine_provider = %s
        os_image = %s
        os_version = %s
        compute_type = %s
        size = %s
        storage_type = %s
	}`, spaceID, nameMbp, machineRoles, machineProvider, osImage, osVersion, computeType, size, storageType)
}

func TestCaasMachineBlueprintCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(testCaasMachineBlueprintDestroy("hpegl_caas_machine_blueprint.test")),
		Steps: []resource.TestStep{
			{
				Config: testCaasMachineBlueprint(),
				Check:  resource.ComposeTestCheckFunc(checkCaasMachineBlueprint("hpegl_caas_machine_blueprint.test")),
			},
		},
	})
}

func TestCaasMachineBlueprintPlan(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testCaasMachineBlueprint(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkCaasMachineBlueprint(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("MachineBlueprint not found: %s", name)
		}
		return nil
	}
}

func testCaasMachineBlueprintDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["hpegl_caas_machine_blueprint.test"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_machine_blueprint.test")
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

		var machineBlueprint *mcaasapi.MachineBlueprint
		machineBlueprints, _, err := p.CaasClient.ClusterAdminApi.V1MachineblueprintsGet(clientCtx, siteIDMbp)
		if err != nil {
			return fmt.Errorf("Error in getting machine blueprint list %w", err)
		}

		for i := range machineBlueprints.Items {
			if machineBlueprints.Items[i].Id == rs.Primary.ID {
				machineBlueprint = &machineBlueprints.Items[i]
			}
		}

		if machineBlueprint != nil {
			return fmt.Errorf("MachineBlueprint still exists")
		}

		return nil
	}
}