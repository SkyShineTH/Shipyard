# Shipyard

A full-stack **GitOps** portfolio project: Go microservices + PostgreSQL, deployed to Kubernetes with **Helm** and **ArgoCD**.

## Architecture (high level)

- **`todo-service`** (Go/Gin + GORM) — REST API for todos
- **`auth-service`** (Go/Gin + JWT + bcrypt) — register/login
- **PostgreSQL** — persistence
- **Helm charts** — `gitops/charts/*`
- **ArgoCD Applications** — `gitops/argocd/*` (GitOps loop)
- **GitHub Actions** — build/push images to GHCR + update Helm image tags

## Repo structure

```text
services/
  todo-service/
  auth-service/
  frontend/               # (planned / WIP depending on timeline)
gitops/
  charts/
    todo-service/
    auth-service/
  argocd/
.github/workflows/
docker-compose.yml
```

## Services & endpoints

### `todo-service` (default port 8080)

- `GET /health`
- `GET /api/v1/todos`
- `POST /api/v1/todos`
- `PUT /api/v1/todos/:id`
- `DELETE /api/v1/todos/:id`

### `auth-service` (default port 8081)

- `GET /health`
- `POST /api/v1/register`
- `POST /api/v1/login`

## Local development (Docker Compose)

Start PostgreSQL + services:

```bash
docker compose up --build
```

Health checks:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8081/health
```

## Kubernetes (kind) + ArgoCD (GitOps)

### Prereqs

- `kubectl`
- `kind`
- ArgoCD installed in the cluster (namespace: `argocd`)

### Deploy via ArgoCD Applications

Apply the Applications:

```bash
kubectl apply -f gitops/argocd/
```

Then watch:

```bash
kubectl -n argocd get applications.argoproj.io
kubectl -n shipyard get deploy,svc,pods
```

### Required secrets (namespace `shipyard`)

These charts load DB/JWT config via `envFrom.secretRef`.

Create secrets:

```bash
kubectl -n shipyard create secret generic todo-service-secret \
  --from-literal=DB_HOST=postgres \
  --from-literal=DB_USER=shipyard \
  --from-literal=DB_PASSWORD=changeme \
  --from-literal=DB_NAME=shipyard \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_SSLMODE=disable

kubectl -n shipyard create secret generic auth-service-secret \
  --from-literal=DB_HOST=postgres \
  --from-literal=DB_USER=shipyard \
  --from-literal=DB_PASSWORD=changeme \
  --from-literal=DB_NAME=shipyard \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_SSLMODE=disable \
  --from-literal=JWT_SECRET="replace-with-a-long-random-secret"
```

### Test APIs on the cluster (port-forward)

```bash
kubectl -n shipyard port-forward svc/shipyard-auth-service 8081:8081
kubectl -n shipyard port-forward svc/shipyard-todo-service 8080:8080
```

Then:

```bash
curl http://127.0.0.1:8081/health
curl http://127.0.0.1:8080/health
```

## CI/CD (GitHub Actions)

On pushes to `main`, workflows will:

- build images
- push to **GHCR**
- update Helm `image.tag` under `gitops/charts/*/values.yaml`
- ArgoCD detects the change and syncs the cluster

## Troubleshooting

### `ImagePullBackOff` / `ErrImagePull`

- usually means image name/tag is wrong or the tag doesn't exist in GHCR
- check the exact image in the Deployment:

```bash
kubectl -n shipyard get deploy shipyard-auth-service -o jsonpath="{.spec.template.spec.containers[0].image}{'\n'}"
kubectl -n shipyard get deploy shipyard-todo-service -o jsonpath="{.spec.template.spec.containers[0].image}{'\n'}"
```

### `CreateContainerConfigError`

- commonly missing secrets or securityContext issues
- check:

```bash
kubectl -n shipyard describe pod <pod-name>
```

### ArgoCD shows `Unknown`

- hard refresh:

```bash
kubectl -n argocd annotate application shipyard-auth-service argocd.argoproj.io/refresh=hard --overwrite
kubectl -n argocd annotate application shipyard-todo-service argocd.argoproj.io/refresh=hard --overwrite
```