# SKE Kubernetes Cluster (2 nodes, g1a.2d)
resource "stackit_ske_cluster" "main" {
  project_id            = var.project_id
  name                  = var.cluster_name
  kubernetes_version_min = var.kubernetes_version

  node_pools = [
    {
      name         = "pool0"
      machine_type = var.machine_type
      minimum      = var.node_count
      maximum      = var.node_count
      max_surge       = 2
      max_unavailable = 0

      os_name       = var.node_image_name
      os_version_min = var.node_image_version

      volume_type = var.volume_type
      volume_size = var.volume_size_gb

      cri = "containerd"

      availability_zones = var.availability_zones
    }
  ]

  maintenance = {
    enable_kubernetes_version_updates    = true
    enable_machine_image_version_updates = true
    start                                = "03:00:00Z"
    end                                  = "04:00:00Z"
  }
}
