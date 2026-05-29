# Shipyard Project Context

## Overview

Shipyard is a full-stack DevOps portfolio project that demonstrates how an
application moves from source code to a working Kubernetes environment through a
GitOps workflow.

- Repository: <https://github.com/SkyShineTH/Shipyard>
- Live demo: <https://shipyard.skyshine.online>
- Public case study: <https://shipyard.skyshine.online/case-study>
- Primary goal: show practical DevOps, Kubernetes, GitOps, and progressive
  delivery skills through a working full-stack app.
- Public positioning: portfolio-grade engineering evidence, not a production
  business service.

## Stack

| Layer | Technology |
| --- | --- |
| Frontend | React, Vite, nginx |
| Backend | Go, Gin, GORM |
| Database | PostgreSQL |
| Containerization | Docker |
| Local runtime | Docker Compose, kind |
| Kubernetes runtime | DigitalOcean Kubernetes for the live demo |
| Packaging | Helm |
| GitOps | Argo CD |
| Progressive delivery | Argo Rollouts |
| CI/CD | GitHub Actions |
| Registry | GHCR |
| Public edge | Cloudflare, DigitalOcean Load Balancer |

## Services

### auth-service

- Go/Gin service for registration and login.
- Uses bcrypt for password hashing.
- Signs JWTs for authenticated frontend/API flows.
- Connects to PostgreSQL through environment variables loaded from a Kubernetes
  Secret.

### todo-service

- Go/Gin service for authenticated todo CRUD operations.
- Uses GORM and PostgreSQL.
- Requires `Authorization: Bearer <JWT>`.
- Deployed as an Argo Rollouts `Rollout` so the project can demonstrate manual
  promotion during canary delivery.

### frontend

- React + Vite app.
- Built into an nginx production image.
- In Kubernetes, nginx serves static assets and proxies `/api/v1/register`,
  `/api/v1/login`, `/api/v1/todos`, and `/api/v1/platform/status` to internal
  services.
- Supports optional origin TLS by mounting a Kubernetes TLS secret into nginx and
  exposing service port `443`.
- Hosts the public `/case-study` route for portfolio evidence.

### platform-status-service

- Go/Gin service for the `/api/v1/platform/status` endpoint.
- Uses an in-cluster Kubernetes ServiceAccount with read-only RBAC.
- Reads only public-safe resource status from the `shipyard` and `argocd`
  namespaces.
- Sanitizes output so public visitors do not see secrets, tokens, pod IPs, node
  names, kubeconfig, database connection strings, or Argo CD credentials.

## GitOps Flow

1. Source changes land on `main`.
2. The relevant GitHub Actions workflow builds the changed service image.
3. The workflow pushes the image to GHCR.
4. The workflow updates the matching Helm chart `image.tag`.
5. Argo CD detects the Git change and syncs the Kubernetes app.
6. Argo Rollouts manages the `todo-service` rollout strategy.
7. The `/case-study` page reads a sanitized infrastructure snapshot from
   `platform-status-service`.

## Live Demo Environment

The current demo runs on DigitalOcean Kubernetes as an on-demand portfolio
environment.

- Cluster name: `shipyard-doks-demo`
- Region: Singapore
- Node pool: `demo-pool`
- Public hostname: `shipyard.skyshine.online`
- TLS mode: Cloudflare to origin HTTPS with a Cloudflare Origin Certificate
- Database: PostgreSQL in the `shipyard` namespace with a small Block Storage PVC
- Entry point: `shipyard-frontend` Service of type `LoadBalancer`

The demo is intentionally tuned for cost control:

- small node footprint
- one frontend replica
- one auth-service replica
- one todo-service replica for the always-on demo state
- one platform-status-service replica for the public read-only snapshot
- reduced CPU requests for app pods
- no high-availability control plane

## Evidence Commands

Useful commands for portfolio proof:

```bash
kubectl -n argocd get applications.argoproj.io
kubectl -n shipyard get pods,svc,pvc -o wide
kubectl -n shipyard get rollout shipyard-todo-service
curl -I https://shipyard.skyshine.online/
curl -s https://shipyard.skyshine.online/api/v1/platform/status
```

Do not publish output that contains API tokens, database passwords, JWT secrets,
kubeconfig content, Cloudflare private keys, or GitHub PATs.

## Milestones

- Built `todo-service` with Go, Gin, GORM, PostgreSQL, Docker, and Compose.
- Built `auth-service` with JWT authentication and bcrypt password hashing.
- Added Helm charts for auth, todo, and frontend.
- Added `platform-status-service` for a sanitized public Kubernetes/GitOps
  snapshot.
- Added Argo CD Application manifests.
- Added GitHub Actions workflows for image build, GHCR push, and chart tag bump.
- Added React/Vite frontend and nginx production serving.
- Added Argo Rollouts for `todo-service`.
- Deployed the full stack to DOKS.
- Added Cloudflare-backed HTTPS for `shipyard.skyshine.online`.
- Added `/case-study` as a public evidence-based portfolio page.
- Captured portfolio evidence in `docs/doks-live-demo.md` and
  `docs/screenshots/`.

## Operational Notes

- Keep `.env`, `.secrets/`, kubeconfig files, tokens, and private keys out of
  Git.
- Use the same `JWT_SECRET` in `auth-service-secret` and `todo-service-secret`.
- If `todo-service` appears `Suspended` in Argo CD, check whether it is paused at
  a canary step and promote the rollout if appropriate.
- If pods are stuck in `Pending` with `Insufficient cpu`, either lower requests
  for the demo or temporarily scale the node pool up.
- Scale the demo down when it is not needed; see `docs/cost-control.md`.
