# Copyright 2021 Hewlett Packard Enterprise Development LP

variable "cluster_name" {
  type        = string
  description = "The name of the cluster"
  default     = "iac-test-clus-21"

  validation {
    condition     = length(var.cluster_name) <= 16
    error_message = "The cluster name must be <= 16 characters in length."
  }
}