# Copyright 2022 Hewlett Packard Enterprise Development LP

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
    api_url = "https://mcaas.intg.hpedevops.net/mcaas"
  }
}

data "hpegl_caas_site" "blr" {
  name = "BLR"
  space_id = "aadc51f2-8b3f-4ae0-aff2-820f7169447f"
}

data "hpegl_caas_cluster_provider" "clusterprovider" {
  name = "ecp"
  site_id = data.hpegl_caas_site.blr.id
}

output "cluster_provider" {
  description = "Storage classes information"
  value       = data.hpegl_caas_cluster_provider.clusterprovider.*
}
