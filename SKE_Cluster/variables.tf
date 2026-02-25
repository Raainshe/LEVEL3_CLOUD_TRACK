# -----------------------------------------------------------------------------
# Stackit Provider & Authentication
# -----------------------------------------------------------------------------

variable "stackit_region" {
  description = "Default region for Stackit resources (e.g. eu01, eu11)"
  type        = string
  default     = "eu01"
}

variable "service_account_key_path" {
  description = "Path to the Stackit service account key JSON file"
  type        = string
}

# -----------------------------------------------------------------------------
# SKE Kubernetes Cluster
# -----------------------------------------------------------------------------

variable "project_id" {
  description = "Stackit project ID"
  type        = string
}

variable "cluster_name" {
  description = "Name of the SKE cluster (lowercase, alphanumeric, hyphens, max 11 chars)"
  type        = string
  default     = "ske-cluster"
}

variable "kubernetes_version" {
  description = "Kubernetes version (e.g. 1.32.11, 1.33.7)"
  type        = string
  default     = "1.33.7"
}

variable "node_count" {
  description = "Number of nodes in the node pool"
  type        = number
  default     = 2
}

variable "machine_type" {
  description = "VM flavor for nodes (e.g. g1a.2d = 2 vCPU, 8 GB)"
  type        = string
  default     = "g1a.2d"
}

variable "node_image_name" {
  description = "OS image for nodes (flatcar or ubuntu)"
  type        = string
  default     = "ubuntu"
}

variable "node_image_version" {
  description = "OS image version (e.g. 2204.20250728.0 for Ubuntu)"
  type        = string
  default     = "2204.20250728.0"
}

variable "volume_type" {
  description = "Storage type for node volumes"
  type        = string
  default     = "storage_premium_perf1"
}

variable "volume_size_gb" {
  description = "Storage size in GB per node"
  type        = number
  default     = 20
}

variable "availability_zones" {
  description = "Availability zones for the node pool"
  type        = list(string)
  default     = ["eu01-1", "eu01-2"]
}
