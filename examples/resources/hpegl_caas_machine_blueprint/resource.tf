# Copyright 2020 Hewlett Packard Enterprise Development LP

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

resource hpegl_caas_machine_blueprint test {

 name = "mbp-test1"
 site_id = "3ad9c737-5bb6-430c-9772-3a6f5a7e4015"
 machine_roles = ["controlplane"]
 machine_provider = "vmaas"
 os_image = "sles-custom"
 os_version = "15"
 compute_type = "General Purpose"
 size = "Large"
 storage_type = "General Purpose"

}