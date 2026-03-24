# Shipyard
A full-stack GitOps platform — React frontend, Go microservices, deployed on Kubernetes via ArgoCD and Helm

## Day 1 (todo-service)

Run local stack (PostgreSQL + todo-service):

```bash
docker compose up --build
```

Service endpoints:

- `GET /health`
- `GET /api/v1/todos`
- `POST /api/v1/todos`
- `PUT /api/v1/todos/:id`
- `DELETE /api/v1/todos/:id`
