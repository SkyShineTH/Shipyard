# Shipyard Cost Control

Shipyard is designed as a portfolio live demo, not an always-on production
system. The goal is to show a real Kubernetes/GitOps workflow while keeping the
DigitalOcean bill predictable.

## Active Resources

The live DOKS demo currently uses:

- 1 DOKS worker node in the `demo-pool` node pool.
- 1 DigitalOcean Load Balancer for the frontend.
- 1 DigitalOcean Block Storage PVC for PostgreSQL.
- Cloudflare DNS/TLS in front of the public hostname.

## Why The Demo Uses One Node

The first deployment hit `Insufficient cpu` on the small shared-CPU node. Instead
of scaling out immediately, the demo was tuned to fit the portfolio workload:

- Reduced app CPU requests for auth-service, todo-service, and frontend.
- Reduced `todo-service` to 1 replica for the always-on demo footprint.
- Kept Argo Rollouts enabled so the canary flow can still be demonstrated.
- Kept PostgreSQL in-cluster for a self-contained demo.

This tradeoff is intentional: it keeps the environment inexpensive while still
showing Kubernetes, Helm, Argo CD, Argo Rollouts, GHCR, and DOKS in one working
system.

## Scale Down After Demo

If the demo is not needed, scale the worker node pool down:

```powershell
doctl kubernetes cluster node-pool update shipyard-doks-demo demo-pool --count 0
```

Before leaving the environment down for a long time, also consider whether to
delete resources that can continue billing independently:

```powershell
kubectl -n shipyard delete svc shipyard-frontend
kubectl -n shipyard delete pvc data-postgres-0
```

Deleting the PVC removes PostgreSQL data. Only do this when the demo data is
disposable.

## Bring The Demo Back

Scale the node pool back up:

```powershell
doctl kubernetes cluster node-pool update shipyard-doks-demo demo-pool --count 1
kubectl get nodes
```

Then confirm the apps recover:

```powershell
kubectl -n argocd get applications.argoproj.io
kubectl -n shipyard get pods,svc,pvc -o wide
```

If `todo-service` is paused on an Argo Rollouts canary step, promote it:

```powershell
kubectl argo rollouts promote shipyard-todo-service -n shipyard
```

