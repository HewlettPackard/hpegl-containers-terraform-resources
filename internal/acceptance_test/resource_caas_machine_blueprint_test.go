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
	nameMbp         = "test-machine-bp"
	machineProvider = "vmaas"
	osImage         = "sles-custom"
	osVersion       = "15"
	computeType     = "General Purpose"
	size            = "G1-CN-xLarge"
	storageType     = "General Purpose"
	apiURLMBp       = "https://mcaas.intg.hpedevops.net/mcaas"
	siteNameMbp     = "FTC"
	workerType      = "Virtual"
	//apiURLMBp       = "https://mcaas.us1.greenlake-hpe.com/mcaas"
)

var machineRoles = []string{"worker"}

// nolint: gosec
func testCaasMachineBlueprint() string {
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
	data "hpegl_caas_site" "blr" {
		name = "%s"
		space_id = var.HPEGL_SPACE
	  }
	resource hpegl_caas_machine_blueprint testmb {
		name         = "%s%d"
  		site_id = data.hpegl_caas_site.blr.id
  		machine_roles = %q
		machine_provider = "%s"
		os_image = "%s"
		os_version = "%s"
		compute_type = "%s"
		size = "%s"
		storage_type = "%s"
        worker_type = "%s"
	}`, apiURLMBp, siteNameMbp, nameMbp, r.Int63n(99999999), machineRoles, machineProvider, osImage, osVersion, computeType, size, storageType, workerType)
}

func TestCaasMachineBlueprintCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              resource.ComposeTestCheckFunc(testCaasMachineBlueprintDestroy("hpegl_caas_machine_blueprint.testmb")),
		Steps: []resource.TestStep{
			{
				Config: testCaasMachineBlueprint(),
				Check:  resource.ComposeTestCheckFunc(checkCaasMachineBlueprint("hpegl_caas_machine_blueprint.testmb")),
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
		rs, ok := s.RootModule().Resources["hpegl_caas_machine_blueprint.testmb"]
		if !ok {
			return fmt.Errorf("Resource not found: %s", "hpegl_caas_machine_blueprint.testmb")
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

		var machineBlueprint *mcaasapi.MachineBlueprint
		field := "applianceID eq " + siteID
		machineBlueprints, _, err := p.CaasClient.MachineBlueprintsApi.V1MachineblueprintsGet(clientCtx, field)
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
