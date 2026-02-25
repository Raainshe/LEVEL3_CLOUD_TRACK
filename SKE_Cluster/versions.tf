terraform {
  cloud {
    organization = "rmakoni"

    workspaces {
      name = "ske-cluster-paas"
    }
  }

  required_version = ">= 1.0"

  required_providers {
    stackit = {
      source  = "stackitcloud/stackit"
      version = ">= 0.14"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.9"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.20"
    }
    local = {
      source  = "hashicorp/local"
      version = ">= 2.4"
    }
  }
}
