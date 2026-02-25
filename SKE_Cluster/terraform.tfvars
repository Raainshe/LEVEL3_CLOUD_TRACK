stackit_region           = "eu01"
service_account_key_path = "key.json"

# SKE Kubernetes Cluster
project_id          = "0436357b-2727-4182-a344-a750ec30744d"
cluster_name        = "ryancluster"
kubernetes_version  = "1.34.3"
node_count          = 2
machine_type        = "g2i.2"
node_image_name     = "ubuntu"
node_image_version  = "2204.20250728.0"
volume_type         = "storage_premium_perf1"
volume_size_gb      = 50
availability_zones  = ["eu01-1", "eu01-2"]