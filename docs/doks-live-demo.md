# Shipyard DOKS Live Demo

Shipyard is deployed as an on-demand portfolio demo on DigitalOcean Kubernetes.
The environment is intentionally small so it can show a real GitOps/Kubernetes
workflow while keeping cloud cost controlled.

## Live Demo

- URL: <https://shipyard.skyshine.online/>
- Repository: <https://github.com/SkyShineTH/Shipyard>
- Cluster: DigitalOcean Kubernetes, Singapore region
- Runtime: 1 shared-CPU worker node for demo use
- Public entry point: DigitalOcean Load Balancer
- TLS: Cloudflare to origin using a Cloudflare Origin Certificate

## What The Demo Shows

- React/Vite frontend served by nginx.
- Go microservices for authentication and todo APIs.
- PostgreSQL running in the cluster with a DigitalOcean Block Storage PVC.
- Helm charts for frontend, auth-service, and todo-service.
- Argo CD Applications managing the deployment from Git.
- Argo Rollouts managing the todo-service progressive delivery flow.
- GitHub Actions building service images, pushing to GHCR, and updating chart image tags.

## Request Flow

```text
User
  -> Cloudflare HTTPS
  -> DigitalOcean Load Balancer
  -> frontend nginx
  -> /api/v1/register, /api/v1/login -> auth-service
  -> /api/v1/todos -> todo-service
  -> PostgreSQL
```

## Delivery Flow

```text
GitHub Actions
  -> build Docker images
  -> push images to GHCR
  -> update Helm chart image tags
  -> Argo CD syncs the target cluster
  -> Argo Rollouts controls todo-service rollout promotion
```

## Current Proof Commands

These commands were captured from the live DOKS environment. They intentionally
avoid printing secrets, tokens, kubeconfig, database passwords, or private keys.

```text
$ kubectl -n argocd get applications.argoproj.io
NAME                    SYNC STATUS   HEALTH STATUS
argo-rollouts           Synced        Healthy
shipyard-auth-service   Synced        Healthy
shipyard-frontend       Synced        Healthy
shipyard-todo-service   Synced        Healthy
```

```text
$ kubectl -n shipyard get pods,svc,pvc -o wide
pod/postgres-0                               1/1   Running
pod/shipyard-auth-service-...               1/1   Running
pod/shipyard-frontend-...                   1/1   Running
pod/shipyard-todo-service-...               1/1   Running

service/shipyard-frontend   LoadBalancer   104.248.98.247   80/TCP,443/TCP
persistentvolumeclaim/data-postgres-0      Bound            2Gi
```

## Screenshots

- [Live demo screenshot](./screenshots/shipyard-live-demo.png)

Additional useful screenshots for portfolio proof:

- Argo CD Applications page showing all apps `Synced` and `Healthy`.
- DOKS cluster page showing `shipyard-doks-demo` and the `demo-pool` node pool.
- Argo Rollouts view or CLI output after promoting `shipyard-todo-service`.
- `kubectl -n shipyard get pods,svc,pvc -o wide` output.

Do not publish screenshots that expose API tokens, JWT secrets, database
passwords, kubeconfig contents, GitHub PATs, or private TLS keys.

