# Shipyard

A full-stack **GitOps** portfolio project: **React + Vite** frontend, **Go** microservices + **PostgreSQL**, deployed to Kubernetes with **Helm** and **ArgoCD**. CI builds images, pushes to **GHCR**, and bumps chart image tags for the GitOps loop.

## Architecture (high level)

- **`todo-service`** (Go/Gin + GORM) — REST API for todos (JWT-protected)
- **`auth-service`** (Go/Gin + JWT + bcrypt) — register / login
- **`frontend`** (React + Vite, nginx in production) — UI; nginx reverse-proxies `/api/v1/*` to the Go services
- **PostgreSQL** — persistence
- **Helm charts** — `gitops/charts/{todo-service,auth-service,frontend}`
- **ArgoCD Applications** — `gitops/argocd/*.yaml` (GitOps loop)
- **GitHub Actions** — `ci-todo.yml`, `ci-auth.yml`, `ci-frontend.yml`: build → push GHCR → update `image.tag` in chart `values.yaml`

## Repo structure

```text
services/
  todo-service/
  auth-service/
  frontend/                 # React app, nginx.conf (Compose), Dockerfile
gitops/
  charts/
    todo-service/
    auth-service/
    frontend/               # ConfigMap nginx for K8s upstreams; Deployment + Service + Ingress
  argocd/
    todo-app.yaml           # Application: shipyard-todo-service
    auth-app.yaml           # Application: shipyard-auth-service
    frontend-app.yaml       # Application: shipyard-frontend
.github/workflows/
docker-compose.yml
CONTEXT.md                  # extended project notes & timeline
```

## Stack (summary)

| Layer        | Technology                          |
| ------------ | ----------------------------------- |
| Frontend     | React 19, Vite 8, react-router-dom  |
| Backend      | Go, Gin, GORM                       |
| Database     | PostgreSQL 16                       |
| Images       | Docker (multi-stage)                |
| Cluster      | kind (local), DOKS (planned)        |
| GitOps       | Helm + ArgoCD                       |
| Registry     | GHCR                                |

## Services & endpoints

### `todo-service` (default port **8080**)

- `GET /health`
- `GET /api/v1/todos` — requires `Authorization: Bearer <JWT>`
- `POST /api/v1/todos`
- `PUT /api/v1/todos/:id`
- `DELETE /api/v1/todos/:id`

### `auth-service` (default port **8081**)

- `GET /health`
- `POST /api/v1/register`
- `POST /api/v1/login`

### `frontend`

- **Docker Compose:** host port **3000** → container **80**; nginx proxies `/api/v1/*` to `auth-service:8081` and `todo-service:8080` on the Compose network.
- **Vite dev:** `npm run dev` (see `services/frontend`); API proxy is configured in `vite.config.js`.
- **Kubernetes:** chart mounts a **ConfigMap** with nginx config; set `upstream.*.host` in `gitops/charts/frontend/values.yaml` to match your backend **Service** names (defaults align with ArgoCD Application names: `shipyard-auth-service`, `shipyard-todo-service`).

## Local development (Docker Compose)

Copy and adjust env files (root `.env` from `.env.example`, etc.), then:

```bash
docker compose up --build
```

Set a non-empty **`JWT_SECRET`** in `.env` so todo and auth agree on token signing (Compose warns if it is missing).

Health checks:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8081/health
curl -I http://127.0.0.1:3000/
```

Frontend-only (from `services/frontend`):

```bash
npm ci
npm run dev
```

## Kubernetes (kind) + ArgoCD (GitOps)

### Prereqs

- `kubectl`
- `kind`
- ArgoCD installed in the cluster (namespace: `argocd`)

### Deploy via ArgoCD Applications

Apply the Applications (adjust `repoURL` in the manifests if you fork):

```bash
kubectl apply -f gitops/argocd/
```

Then watch:

```bash
kubectl -n argocd get applications.argoproj.io
kubectl -n shipyard get deploy,svc,pods
```

### Required secrets (namespace `shipyard`)

Todo and auth charts load DB (and todo loads JWT) via `envFrom.secretRef`. **Use the same `JWT_SECRET` for both** `auth-service-secret` and `todo-service-secret`.

```bash
kubectl -n shipyard create secret generic todo-service-secret \
  --from-literal=DB_HOST=postgres \
  --from-literal=DB_USER=shipyard \
  --from-literal=DB_PASSWORD=changeme \
  --from-literal=DB_NAME=shipyard \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_SSLMODE=disable \
  --from-literal=JWT_SECRET="replace-with-a-long-random-secret"

kubectl -n shipyard create secret generic auth-service-secret \
  --from-literal=DB_HOST=postgres \
  --from-literal=DB_USER=shipyard \
  --from-literal=DB_PASSWORD=changeme \
  --from-literal=DB_NAME=shipyard \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_SSLMODE=disable \
  --from-literal=JWT_SECRET="replace-with-a-long-random-secret"
```

### Test on the cluster (port-forward)

```bash
kubectl -n shipyard port-forward svc/shipyard-frontend 3000:80
kubectl -n shipyard port-forward svc/shipyard-auth-service 8081:8081
kubectl -n shipyard port-forward svc/shipyard-todo-service 8080:8080
```

Then open `http://127.0.0.1:3000` or hit the APIs directly, for example:

```bash
curl http://127.0.0.1:8081/health
curl http://127.0.0.1:8080/health
```

## CI/CD (GitHub Actions)

On pushes to **`main`** (path-filtered per service), workflows:

- build the Docker image for that service
- push to **GHCR** (`ghcr.io/<owner>/shipyard-<service>`)
- commit an update to `gitops/charts/<service>/values.yaml` (`image.tag`, `[skip ci]`)

ArgoCD picks up the Git change and syncs the cluster.

| Workflow        | Paths (among others)                          |
| --------------- | ---------------------------------------------- |
| `ci-todo.yml`   | `services/todo-service/**`, chart todo         |
| `ci-auth.yml`   | `services/auth-service/**`, chart auth         |
| `ci-frontend.yml` | `services/frontend/**`, chart frontend       |

## Troubleshooting

### `ImagePullBackOff` / `ErrImagePull`

- Image name/tag wrong or tag not yet in GHCR; confirm the Deployment image:

```bash
kubectl -n shipyard get deploy shipyard-auth-service -o jsonpath="{.spec.template.spec.containers[0].image}{'\n'}"
kubectl -n shipyard get deploy shipyard-todo-service -o jsonpath="{.spec.template.spec.containers[0].image}{'\n'}"
kubectl -n shipyard get deploy shipyard-frontend -o jsonpath="{.spec.template.spec.containers[0].image}{'\n'}"
```

### `CreateContainerConfigError`

- Often missing secrets or volume/security context issues:

```bash
kubectl -n shipyard describe pod <pod-name>
```

### Frontend API proxy fails in Kubernetes

- Ensure `gitops/charts/frontend/values.yaml` → `upstream.auth.host` / `upstream.todo.host` match the **Service** names of auth and todo in `shipyard` (defaults: `shipyard-auth-service`, `shipyard-todo-service`).

### ArgoCD shows `Unknown`

- Hard refresh:

```bash
kubectl -n argocd annotate application shipyard-auth-service argocd.argoproj.io/refresh=hard --overwrite
kubectl -n argocd annotate application shipyard-todo-service argocd.argoproj.io/refresh=hard --overwrite
kubectl -n argocd annotate application shipyard-frontend argocd.argoproj.io/refresh=hard --overwrite
```

## Conventions (short)

- **K8s namespace:** `shipyard`
- **ArgoCD app names:** `shipyard-todo-service`, `shipyard-auth-service`, `shipyard-frontend`
- **Images:** `ghcr.io/<github-username-lowercase>/shipyard-{todo-service,auth-service,frontend}:<tag>`
- **GitOps branch:** `main`

See [CONTEXT.md](./CONTEXT.md) for timeline, conventions, and detailed notes.
