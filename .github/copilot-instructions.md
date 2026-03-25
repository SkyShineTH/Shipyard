# Project Guidelines

## Scope And Source Of Truth
- This repository is in bootstrap stage. Treat [CONTEXT.md](CONTEXT.md) as the primary planning and architecture source.
- Keep instructions minimal and actionable. Link to existing docs instead of duplicating large sections.

## Architecture
- Target architecture is a GitOps monorepo with three services:
  - `services/todo-service` (Go, Gin, GORM, PostgreSQL)
  - `services/auth-service` (Go, Gin, JWT, PostgreSQL)
  - `services/frontend` (React + Vite, nginx image)
- GitOps assets are expected under `gitops/charts` and `gitops/argocd`.
- CI workflows are expected under `.github/workflows`.

## Build And Test
- There are currently no runnable service directories or test suites in this workspace.
- Before running build/test commands, first verify files exist and prefer discovery over assumptions.
- When scaffolding starts, prefer these defaults unless project files specify otherwise:
  - Go services: `go test ./...` and `go build ./...`
  - Frontend: `npm test` (if configured), `npm run build`, `npm run dev`
  - Local stack: `docker compose up`

## Conventions
- Keep naming and deployment conventions aligned with [CONTEXT.md](CONTEXT.md):
  - Go module path: `github.com/SkyShineTH/shipyard/<service-name>`
  - Docker image: `ghcr.io/SkyShineTH/shipyard-<service-name>:<tag>`
  - Namespace: `shipyard`
  - ArgoCD app name: `shipyard-<service-name>`
- `main` is the production branch watched by ArgoCD.

## Coding Guidance
- For Go files, add comments only for Go-specific patterns (error handling decisions, struct tags, interfaces) and avoid line-by-line narration.
- Prefer small, focused changes that preserve planned service boundaries.
- If repo structure in [CONTEXT.md](CONTEXT.md) and actual files diverge, follow existing code in the workspace and update docs in the same change when appropriate.

## Documentation
- Keep [README.md](README.md) concise and high level.
- Put detailed implementation and rollout details in dedicated docs and link from README.
