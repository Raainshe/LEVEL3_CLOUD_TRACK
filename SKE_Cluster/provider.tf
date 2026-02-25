# Stackit Provider Configuration
provider "stackit" {
  default_region           = var.stackit_region
  service_account_key_path = var.service_account_key_path
}

# Kubernetes Provider Configuration
provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Helm Provider Configuration
provider "helm" {
  kubernetes = {
    config_path = "~/.kube/config"
  }
}