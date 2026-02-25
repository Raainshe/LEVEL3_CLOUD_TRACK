# rmakoni_lvl3Cloud

A project repository for Level 3 Cloud computing coursework and OpenStack exploration. It also contains deployment documentation for a Kubernetes-based Platform-as-a-Service (PAAS) built for this course.

## Platform-as-a-Service (PAAS)

A PAAS that runs on a Kubernetes cluster. The cluster setup is in [SKE_Cluster/](SKE_Cluster/).

- **Stack**: Vue front end, Go backend using Gin.
- **Features**:
  - Provision, update, and delete Redis instances
  - Connection link to access each Redis instance
  - Service logs and audit logs
  - Admin capabilities to manage everything
- **Repo layout**: [PAAS/](PAAS/) holds the application; [SKE_Cluster/](SKE_Cluster/) holds the cluster setup.

## Deployment documentation

This section documents **how everything was deployed** (cluster and PAAS), not study material.

### Kubernetes cluster

Setup and provisioning are in [SKE_Cluster/](SKE_Cluster/) (e.g. STACKIT SKE).

### PAAS on the cluster

What is deployed and where the manifests live:

**Frontend** — [PAAS/frontend/deployment/](PAAS/frontend/deployment/):

- `frontend-deployment.yaml`, `frontend-service.yaml` — app workload and service
- `frontend-ingress.yaml` — NGINX Ingress, TLS (e.g. ryanpaas.stackit.gg), `/api` to backend, `/` to frontend
- `cert-manager-issuer.yaml` — Let's Encrypt issuer for TLS
- `frontend-servicemonitor.yaml` — Prometheus scraping (`/metrics`, 15s)
- `frontend-rbac.yaml` — ServiceAccount for frontend

**Backend** — [PAAS/backend/](PAAS/backend/):

- `backend-deployment.yaml`, `backend-service.yaml` — app workload and service
- `backend-servicemonitor.yaml` — Prometheus scraping
- `backend-rbac.yaml` — ServiceAccount, Role/ClusterRole for Redis (redisfailovers), namespace creation; ClusterRoleBinding

The cluster uses NGINX Ingress, cert-manager, and Prometheus (kube-prometheus-stack) as reflected in these manifests.

## Week 1

### Resources

- [What is OpenStack?](https://www.notion.so/What-is-OpenStack-2edfebed13d0800a8859d9b39296f71d?source=copy_link)
- [Creating an Instance](https://www.notion.so/Creating-an-Instance-2effebed13d08074ba08ef037b89e3bc?source=copy_link)
- [Single-Node DevStack Architecture Diagram](https://www.figma.com/design/kTpV5hDLGfyuKH9orza67s/Single-Node-DevStack-Architecture-Diagram?m=auto&t=Pv7tVfEpuWV7hqgH-1)

## Week 2

### Resources

- [Terraform](https://www.notion.so/Terraform-2f4febed13d080428ae2f59d731a5e7a?source=copy_link)
- [Kubernetes](https://www.notion.so/Kubernettes-2f4febed13d080b1b4e3f44bc5c7e80a?source=copy_link)

## Week 3

## Resources
- [STACKIT SKE](https://www.notion.so/STACKIT-SKE-Terraform-2fbfebed13d08063874fc2f7145ce23a?source=copy_link)

## Week 4

### Resources
- [API](https://www.notion.so/API-303febed13d0803a9576d1b893407f2c?source=copy_link)

## Week 5

### Resources
- [Ingress Controller](https://www.notion.so/Ingress-Controller-30bfebed13d080b69514d5750bd9a5e6?source=copy_link)
- [Frontend](https://www.notion.so/Frontend-30cfebed13d080198e93c84d52aa701c?source=copy_link)

## Week 6

### Resources
- [Monitoring](https://www.notion.so/Monitoring-312febed13d08098bcd9fb5a1581570d?source=copy_link)
