output "cluster_name" {
  description = "Name of the SKE cluster"
  value       = stackit_ske_cluster.main.name
}

output "cluster_id" {
  description = "ID of the SKE cluster"
  value       = stackit_ske_cluster.main.id
}
