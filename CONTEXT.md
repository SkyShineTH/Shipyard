# Shipyard — Project Context

## Overview
**Name:** Shipyard  
**Repo:** `https://github.com/SkyShineTH/Shipyard`  
**Description:** A full-stack GitOps platform — React frontend, Go microservices, deployed on Kubernetes via ArgoCD and Helm  
**Goal:** Fullstack DevOps portfolio — โชว์ทั้ง frontend + backend + infra ครบ  
**Owner:** นักศึกษาจบใหม่ สาย DevOps มีประสบการณ์ฝึกงาน DevOps มาแล้ว

---

## Stack

| Layer | Technology |
|---|---|
| Frontend | React + Vite |
| Backend | Go (Gin framework + GORM) |
| Database | PostgreSQL |
| Container | Docker (multi-stage build) |
| Orchestration | Kubernetes |
| Local K8s | kind (Docker-based) |
| Production K8s | DigitalOcean Kubernetes (DOKS) — เคย deploy แล้ว **destroy cluster** แล้ว (ไม่มีสภาพแวดล้อม DO สาธารณะ) |
| Package manager | Helm |
| GitOps controller | ArgoCD |
| Progressive delivery | Argo Rollouts (canary deployment) |
| CI/CD | GitHub Actions |
| Registry | GHCR (GitHub Container Registry) |

---

## Services

### 1. `todo-service` (Go)
- Framework: Gin
- ORM: GORM
- DB: PostgreSQL
- Endpoints: `GET /api/v1/todos`, `POST /api/v1/todos`, `PUT /api/v1/todos/:id`, `DELETE /api/v1/todos/:id`
- Health: `GET /health`
- Port: `8080`

### 2. `auth-service` (Go)
- Framework: Gin
- Auth: JWT
- DB: PostgreSQL
- Endpoints: `POST /api/v1/register`, `POST /api/v1/login`
- Health: `GET /health`
- Port: `8081`

### 3. `frontend` (React + Vite)
- Calls: todo-service และ auth-service APIs
- Served by: nginx (production Docker image)
- Port: `3000` (dev), `80` (container)

---

## Repo Structure

```
shipyard/
├── services/
│   ├── todo-service/
│   │   ├── main.go
│   │   ├── handler/todo.go
│   │   ├── model/todo.go
│   │   ├── db/db.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── .env.example
│   ├── auth-service/
│   │   ├── main.go
│   │   ├── handler/auth.go
│   │   ├── model/user.go
│   │   ├── middleware/jwt.go
│   │   ├── db/db.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── .env.example
│   └── frontend/
│       ├── src/
│       ├── Dockerfile
│       ├── nginx.conf
│       └── .env.example
├── gitops/
│   ├── charts/
│   │   ├── todo-service/
│   │   ├── auth-service/
│   │   └── frontend/
│   └── argocd/
│       ├── argo-rollouts-app.yaml
│       ├── todo-app.yaml
│       ├── auth-app.yaml
│       └── frontend-app.yaml
├── .github/
│   └── workflows/
│       ├── ci-todo.yml
│       ├── ci-auth.yml
│       └── ci-frontend.yml
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Timeline (2 สัปดาห์)

### Week 1 — Backend + Infrastructure
| Day | งาน |
|---|---|
| Day 1 | todo-service (Go): boilerplate, Dockerfile, docker-compose |
| Day 2 | auth-service (Go): JWT, boilerplate, Dockerfile |
| Day 3 | Helm charts: todo + auth (Chart.yaml, templates/, values.yaml) |
| Day 4 | GitHub Actions CI: build → push GHCR → update image tag |
| Day 5 | kind cluster + ArgoCD install + Application manifests + GitOps loop ทดสอบ |

### Week 2 — Frontend + Production
| Day | งาน |
|---|---|
| Day 6 | React frontend: pages, API calls, nginx Dockerfile |
| Day 7 | Helm chart สำหรับ frontend + CI workflow |
| Day 8 | DigitalOcean DOKS cluster + deploy ทั้ง 3 services |
| Day 9 | Argo Rollouts: canary deployment (20% → 50% → 100%) |
| Day 10 | README, architecture diagram, live demo URL, GitHub badges |

**ลำดับที่ทำจริง (ยืดหยุ่น):** ทำ **Day 9 → Day 10** บน **kind** ก่อนได้ — Rollouts ไม่บังคับ DOKS; **Day 10** ทำ README/diagram/badges ก่อน แล้วค่อยเติม **live demo URL** หลัง **Day 8** เมื่อมี DOKS

---

## Infrastructure Notes

- **GitOps pattern:** แยก app-repo (source code) กับ gitops config ไว้ใน folder `gitops/` ของ repo เดียวกัน
- **Image tag update:** GitHub Actions push image แล้ว auto-update `image.tag` ใน `values.yaml` → ArgoCD detect diff → sync อัตโนมัติ
- **Canary steps:** 20% → pause (manual promote) → 50% → 100%
- **Local dev:** `docker compose up` รัน 3 services + PostgreSQL พร้อมกัน
- **DOKS:** เคยใช้ credit / cluster สำหรับการเรียนรู้ — cluster ถูก destroy แล้วเพื่อลดค่าใช้จ่าย; คู่มือติดตั้งยังอยู่ใน README

---

## Conventions

- **Go module path:** `github.com/<YOUR_USERNAME>/shipyard/<service-name>`
- **Docker image:** `ghcr.io/<YOUR_USERNAME>/shipyard-<service-name>:<tag>`
- **Helm release name:** `<service-name>` (e.g. `todo-service`)
- **K8s namespace:** `shipyard`
- **ArgoCD app name:** `shipyard-<service-name>`
- **Branch:** `main` เป็น production branch ที่ ArgoCD watch

---

## Go Notes (สำหรับ agent)
- ผู้พัฒนาเคยอ่าน Go tutorial มาบ้าง ยังไม่คล่อง
- ใส่ comment อธิบายเฉพาะ Go-specific patterns เช่น error handling, struct tags, interface — ไม่ต้องอธิบายทุกบรรทัด
- ใช้ Gin สำหรับ HTTP router, GORM สำหรับ ORM, godotenv สำหรับ .env

---

## Current Status
- [x] ชื่อโปรเจค: Shipyard
- [x] Stack confirmed
- [x] Repo structure confirmed
- [x] Timeline confirmed
- [x] Day 1: todo-service — boilerplate + Dockerfile + docker-compose
- [x] Day 2: auth-service — JWT, boilerplate (model/db/middleware/handler), Dockerfile, docker-compose
- [x] Day 3: Helm charts — todo-service + auth-service (Chart.yaml, values.yaml, deployment/service/ingress templates)
- [x] Day 4: GitHub Actions CI: build → push GHCR → update image tag (todo + auth)
- [x] Day 5: kind cluster, ArgoCD install, Application manifests, GitOps loop test
- [x] Day 6: React frontend — pages (home, todos, login, register), API client, Vite dev proxy, nginx Dockerfile, docker-compose `frontend` service
- [x] Day 7: Helm chart `gitops/charts/frontend` (Deployment, Service, Ingress, ConfigMap nginx for K8s upstreams), GitHub Actions `ci-frontend.yml`, ArgoCD `frontend-app.yaml`
- [x] Day 9: Argo Rollouts — `gitops/argocd/argo-rollouts-app.yaml`; `todo-service` chart uses `Rollout` + canary (20% → pause → 50% → 100%); README badges + diagram + rollouts usage
- [x] DOKS: เคยมี live demo บน DO — cluster ถูก destroy แล้ว; portfolio สาธารณะผ่าน GitHub Pages (`skyshine.online`)