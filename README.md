# hpegl-caas-terraform-resources

- [hpegl-containers-terraform-resources](#hpegl-containers-terraform-resources)
    * [Introduction](#introduction)
    * [Terraform versions](#terraform-versions)
    * [Terraform provider & v2.0 SDK](#terraform-provider--v20-sdk)
    * [Basic structure](#basic-structure)
    * [internal](#internal)
        + [acceptance_test](#acceptance_test)
            - [Running acceptance tests](#running-acceptance-tests)
        + [resources](#resources)
        + [test-utils](#test-utils)
    * [Building and using the "dummy" service-specific provider](#building-and-using-the-dummy-service-specific-provider)
        + [Building the service-specific provider](#building-the-service-specific-provider)
        + [Using the service-specific provider](#using-the-service-specific-provider)
        + [IAM token generation](#iam-token-generation)
            - [API-vended Service Client](#api-vended-service-client)
            - [HPE Service Client](#hpe-service-client)
    * [To Build and Test the Terraform Provider](#to-build-and-test-the-terraform-provider)

## Introduction

This repo contains Containers terraform provider code that we've used while developing
the hpegl provider (https://github.com/HewlettPackard/terraform-provider-hpegl).  It is
the exemplar service provider repo, and will form the basis of a Github template.

## Terraform version

This code has been tested against Terraform version v1.1.0

## Terraform provider & v2.0 SDK

We are mandating the use of the v2.0 Terraform SDK.  This version of the SDK is documented
[here](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#new-diagnostics-enabled-map-validators).  Perhaps
the two most pertinent enhancements are the passing of a context down to the CRUD implementation functions, and the
ability to bundle errors and warnings into a Diagnostics list ([]Diagnostic) that is returned from the CRUD functions.
These Diagnostics can be warnings or errors, and are processed by Terraform for presentation on the console.
The Terraform provider writing tutorial [here](https://learn.hashicorp.com/collections/terraform/providers) has been updated
to use the v2.0 SDK.  

## Basic structure

The repo follows golang standards:

```bash
.
├── cmd
│   └── terraform-provider-hpegl
├── examples
│   ├── cluster-create
│   ├── resources
│   └── variables.tf
├── golangci-lint-config.yaml
├── go.mod
├── go.sum
├── integ.yaml
├── internal
│   ├── acceptance_test
│   ├── resources
│   └── test-utils
├── main.go
├── Makefile
├── pkg
│   ├── auth
│   ├── client
│   ├── constants
│   └── resources
├── prod.yaml
├── README.md
├── templates
│   └── resources
└── tools
    └── tools.go
```

## internal

The internal directory contains:
* Terraform acceptance tests in acceptance_test
* Service resource terraform CRUD implementation code in resources
* Code in test-utils that returns a plugin.ProviderFunc object that can be used
    to create a "dummy" provider for development purposes. This object can also be used
    for acceptance tests.
  

### internal/acceptance_test

There are two basic types of [terraform test](https://www.terraform.io/docs/extend/testing/index.html):
* [Unit tests](https://www.terraform.io/docs/extend/testing/unit-testing.html): <br>
  Terraform unit-tests are, according to the documentation, used mainly for "testing helper methods that
  expand or flatten API responses into data structures for storage into state by Terraform"


* [Acceptance tests](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html): <br>
  Acceptance tests require real service instances to run against, and exercise CRUD of real service objects.
  These tests are the primary method of ensuring that Terraform providers work as advertised. 
  Acceptance tests need to be developed for each service that is added to the GreenLake provider.
  Acceptance tests can be run from this service provider repo using the plugin.ProviderFunc object
  returned from test-utils. Moreover the intention is that we will be able to copy acceptance tests
  verbatim (or with minimal changes to test set-up) from the service repos to hpegl.  See
  [here](https://github.com/hpe-hcss/terraform-provider-hpegl/tree/main/internal/acceptance/bmaas)
  for acceptance tests for bmaas/Quake.  Note that these bmaas acceptance tests require a working Quake
  portal to run against.  Also note that these bmaas acceptance tests were copied from the standalone
  Quake provider repo. <br><br>
  Some information on writing acceptance tests can be found
  [here](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html)

  This repo implements the acceptance framework. More info on this framework can be found [here](https://github.com/hpe-hcss/terraform-provider-hpegl/blob/main/developer/HPEGL_developer_details.md#acceptance-test-framework)
  
An example of how to use test-utils.ProviderFunc() to populate a test provider map required by acceptance
tests is contained in provider_test.go

#### Running acceptance tests

To run acceptance tests:

```bash
$ make acceptance
```

### internal/resources

This repo contains CaaS provider code to create and destroy a CaaS cluster, along with some stub cluster-blueprint
CRUD code.  This code is ultimately executed by the hpegl terraform provider.  Note that we are using
the v2.0 SDK.  The Terraform provider writing tutorial [here](https://learn.hashicorp.com/collections/terraform/providers) has been updated
to use the v2.0 SDK.  One of the features of the v2.0 SDK is that
a context is passed-down to the CRUD functions which allows terraform to time-out 
operations by cancelling the context.  With regard to setting timeouts, note this map in the cluster resource definition:
```go
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(clusterAvailableTimeout),
			// Update: schema.DefaultTimeout(clusterAvailableTimeout),
			Delete: schema.DefaultTimeout(clusterDeleteTimeout),
		},
```

The timeout settings are used by Terraform itself.  When a timeout is exceeded the context passed-in is cancelled,
and the provider code should handle this.  In our case the CaaS polling code will exit.

A diag.Diagnostics slice is returned
by the CRUD functions which is processed by terraform.  Errors and warnings are presented on the console to
the user.  There are helper functions in the diag library to create errors and warnings, see the code
for examples of how to use them.

The resource code makes use of the client.GetClientFromMetaMap() function to extract the Client object from
the meta argument to each CRUD function.  The meta argument is passed-in by terraform, and in the case
of hpegl is a map[string]interface{} of service clients, where each service client defines the key for its
Client in the map.  The map also contains the Token Retrieve Function.

Note that because this repo defines a service block in the provider stanza which is optional then it is
possible that the entry in the meta map is nil.  Thus GetClientFromMetaMap checks for this and returns
an error which is then passed up to Terraform to be displayed on the console.

The resource code makes use of the client.GetToken() function to extract the Token Retrieve Function from
the meta map and to execute it.

The bulk of service-team development will occur in this directory.  Please ensure that as resource and data-source
terraform objects are created that they are added to pkg/resources/registration.go

### internal/test-utils

The code in this directory should not need to be changed outside of imports.  It constructs a terraform plugin.ProviderFunc object
that is used to create a "dummy" service-specific terraform provider called hpegl that can be used for
development work.  See [later](#building-and-using-the-dummy-service-specific-provider) for information on how to
build and use this "dummy" provider. The "dummy" service-specific provider that is created with this function has all of the provider 
configuration information that is required by the real hpegl provider, which calls-into the same NewProviderFunc()

The code uses the registration.ServiceRegistration and client.InitialiseClient interfaces created in pkg/reources
and pkg/client respectively to create the ProviderFunc.  This means that only the resources and data-sources in
addition to the Client defined in this repo are available to the service-specific terraform provider.

The IAM token Handler is initialised and used to create a Token Retrieve Function that is also passed-in to create
the ProviderFunc.
  
## Building and using the "dummy" service-specific provider

The code that exposes the plugin.ProviderFunc object created in internal/test-utils to terraform as a provider
is contained in cmd/terraform-provider-hpegl and should not need to be altered outside of imports.

The service-specific provider is exposed to terraform with the name "hpegl", the same as that of the overall GreenLake
provider.  This means that the resource definitions in pkg/resources do not have to be changed in any way
when using the service-specific provider in hcl for provider development or in acceptance tests.

### Building the service-specific provider

To build the provider type "make install".  This will build the provider binary and also place it
in a .local directory that is compatible with terraform versions >= v0.13.  It is important to note
that the name of the service under development should be included in the .local directory path.  Note the
following section at the head of the Makefile:

```makefile
NAME=$(shell find cmd -name "main.go" -exec dirname {} \; | sort -u | sed -e 's|cmd/||')
VERSION=0.0.1
DUMMY_PROVIDER=caas
LOCAL_LOCATION=~/.local/share/terraform/plugins/terraform.example.com/$(DUMMY_PROVIDER)/hpegl/$(VERSION)/linux_amd64/
```

The value of DUMMY_PROVIDER must be changed to something specific to the service under development, we suggest
that the value contains the service mnemonic.

See below for how to use this service-specific provider in development.

### Using the service-specific provider

The service specific provider will be exposed to terraform under the name "hpegl".  We need to make sure
that terraform will use the provider stored in the DUMMY_PROVIDER location specified in the Makefile.
To do this note main.tf in examples, in particular the terraform stanza:

```hcl
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
```

Note the value of "source" under hpegl.  It needs to be of the form:
```bash
"terraform.example.com/$(DUMMY_PROVIDER)/hpegl"
```
where $(DUMMY_PROVIDER) is replaced by the value of DUMMY_PROVIDER from the Makefile.  With this
"source" terraform will use the service-specific provider that was generated by "make install".
The "hpegl" provider stanza can be provided and service resources and data-sources can be referred to by
the keys specified in pkg/resources in hcl (i.e. .tf) files for development and testing purposes.

### IAM token generation

We've added support for programmatically-generated IAM tokens.  These tokens are returned by a
Token Retrieve Function that is created and placed in the meta map[string]interface{} by the hpegl
provider.  At the moment we use service client creds to generate tokens.  We support two types of
service-client:

* HPE Service Clients
* API-vended Service Clients

By default we assume that the service client being used is an API-vended one.  The information needed
by the IAM libraries for each type of service client is exposed as env-vars.

#### API-vended Service Client

```bash
export HPEGL_TENANT_ID=< tenant-id >
export HPEGL_USER_ID=< service client id >
export HPEGL_USER_SECRET=< service client secret >
export HPEGL_IAM_SERVICE_URL=< the "issuer" URL for the service client  >
```

#### HPE Service Client

```bash
export HPEGL_TENANT_ID=< tenant-id >
export HPEGL_USER_ID=< service client id >
export HPEGL_USER_SECRET=< service client secret >
export HPEGL_IAM_SERVICE_URL=< GL iam service url, defaults to https://client.greenlake.hpe.com/api/iam >
export HPEGL_API_VENDED_SERVICE_CLIENT=false
```
## To Build and Test the Terraform Provider

Pre-requisites:

* The git command-line utility
* The go programming language package
* Terraform (Version: v1.1.0)

Build the Terraform Provider binary:

```bash
$ cd hpegl-containers-terraform-resources
$ make build
```

Install the Terraform Provider binary to the local env:

```bash
$ make install
```

Export the required environment variables:

```bash
export HPEGL_TENANT_ID=<tenant-id>
export HPEGL_USER_ID=<service client id>
export HPEGL_USER_SECRET=<service client secret>
export HPEGL_IAM_SERVICE_URL=<the "issuer" URL for the service client >
```

To initialize the terraform:

```bash
$ cd examples/resources/hpegl_caas_cluster
$ terraform init
```

Note: Ensure there is no .terraform.lock.hcl or .terraform in examples/cluster-create before running terraform init. If this file/folder is present, manually delete it beform initializing the terraform. 

Update examples/resources/hpegl_caas_cluster/resource.tf with values for

```bash
cluster name
cluster blueprint name
site name
space id
```
To create the terraform plan:

```bash
terraform plan
```

To apply the plan and create a cluster:

```bash
terraform apply
```
Note: The timeout for cluster creation is set to 60 mins.

To delete the cluster:

```bash
terraform destroy
```


