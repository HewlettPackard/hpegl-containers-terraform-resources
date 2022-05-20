# Copyright 2020 Hewlett Packard Enterprise Development LP

terraform {
  required_providers {
    hpegl = {
      # We are specifying a location that is specific to the service under development
      # In this example it is caas (see "source" below).  The service-specific replacement
      # to caas must be specified in "source" below and also in the Makefile as the
      # value of DUMMY_PROVIDER.
      source  = "hpe/hpegl"
      version = ">= 0.2.0"
    }
  }
}

provider hpegl {
  caas {
    api_url = "https://mcaas.intg.hpedevops.net/mcaas/v1"
  }
}

data "hpegl_caas_cluster_blueprint" "bp" {
  name = "demo"
  space_id = ""
}

data "hpegl_caas_site" "blr" {
  name = "BLR"
  space_id = ""
}

resource hpegl_caas_cluster test {
  name         = "tf-test"
  blueprint_id = data.hpegl_caas_cluster_blueprint.bp.id
  site_id = data.hpegl_caas_site.blr.id
  space_id     = ""
}
