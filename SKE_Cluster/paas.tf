
# 1. Kubeconfig Retrieval
# Fetch the kubeconfig for the cluster so we can configure the Helm/K8s providers
resource "stackit_ske_kubeconfig" "current" {
  project_id   = var.project_id
  cluster_name = stackit_ske_cluster.main.name
}

# Write the kubeconfig to a local file (required by the providers)
resource "local_file" "kubeconfig" {
  content  = stackit_ske_kubeconfig.current.kube_config
  filename = "${path.module}/kubeconfig.yaml"
}
# 2. Redis Operator (Platform Software)

# Install the Redis Operator via Helm
resource "helm_release" "redis_operator" {
  name             = "redis-operator"
  repository       = "https://spotahome.github.io/redis-operator"
  chart            = "redis-operator"
  namespace        = "operators"
  create_namespace = true
  version          = "3.2.9"

  # Ensure cluster is ready before trying to install
  depends_on = [stackit_ske_cluster.main, local_file.kubeconfig]
}

# 3. Redis Gateway Namespace (for external access entrypoint)

resource "kubernetes_namespace" "redis_gateway" {
  metadata {
    name = "redis-gateway"
  }

  depends_on = [local_file.kubeconfig]
}

# 4. Redis Gateway LoadBalancer Service
resource "kubernetes_service" "redis_gateway_lb" {
  metadata {
    name      = "redis-gateway-lb"
    namespace = kubernetes_namespace.redis_gateway.metadata[0].name
    labels = {
      app = "redis-gateway"
    }
  }

  spec {
    type = "LoadBalancer"

    selector = {
      app = "redis-gateway"
    }

    port {
      name       = "redis-gateway"
      port       = 6379
      target_port = 6379
      protocol   = "TCP"
    }
  }

  depends_on = [kubernetes_namespace.redis_gateway]
}
