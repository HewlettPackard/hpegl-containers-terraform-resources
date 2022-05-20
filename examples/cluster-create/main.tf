# Copyright 2020 Hewlett Packard Enterprise Development LP

# Set-up for terraform >= v0.13
terraform {
  required_providers {
    hpegl = {
      # We are specifying a location that is specific to the service under development
      # In this example it is caas (see "source" below).  The service-specific replacement
      # to caas must be specified in "source" below and also in the Makefile as the
      # value of DUMMY_PROVIDER.
      source  = "terraform.example.com/caas/hpegl"
      version = ">= 0.0.1"
    }
  }
}

provider hpegl {
  caas {
    api_url = "https://client.greenlake.hpe.com/api/caas/mcaas/v1"
  }
}

resource hpegl_caas_cluster test {
  name         = var.cluster_name
  blueprint_id = ""
  site_id = ""
  space_id     = ""
}
