# Shipyard Monitoring

Shipyard supports an on-demand Prometheus and Grafana stack for portfolio
evidence. The stack is intentionally private: Grafana and Prometheus are
`ClusterIP` services and should be accessed through `kubectl port-forward`, not
through a public Load Balancer.

## What Is Included

- `/metrics` endpoints on `auth-service`, `todo-service`, and
  `platform-status-service`.
- `shipyard-monitoring` Argo CD Application for `kube-prometheus-stack`.
- `shipyard-observability` Argo CD Application for Shipyard `ServiceMonitor`
  resources and a custom Grafana dashboard.
- A Grafana dashboard named `Shipyard HTTP Overview`.

The monitoring stack is meant for demos, screenshots, and interviews. It is not
required for the public app to run.

## Deploy

Push the service changes first so GitHub Actions can build images that include
the `/metrics` endpoints.

Register the public Prometheus Community Helm repository in Argo CD:

```powershell
kubectl apply -f gitops/argocd/prometheus-community-repo.yaml
```

On the one-node demo cluster, the full Prometheus/Grafana stack can exceed the
available CPU and memory. For screenshots, temporarily scale the node pool to 2
nodes, then scale it back down after teardown:

```powershell
doctl kubernetes cluster node-pool update shipyard-doks-demo demo-pool --count 2 --min-nodes 1 --max-nodes 2 --auto-scale
```

Then install the monitoring stack:

```powershell
kubectl apply -f gitops/argocd/monitoring-app.yaml
kubectl -n argocd wait --for=jsonpath="{.status.health.status}"=Healthy application/shipyard-monitoring --timeout=900s
```

After the Prometheus Operator CRDs are available, install the Shipyard-specific
ServiceMonitors and dashboard:

```powershell
kubectl apply -f gitops/argocd/shipyard-observability-app.yaml
kubectl -n argocd wait --for=jsonpath="{.status.health.status}"=Healthy application/shipyard-observability --timeout=300s
```

## Verify

```powershell
kubectl -n monitoring get pods,svc
kubectl -n shipyard get servicemonitor
kubectl -n argocd get applications.argoproj.io shipyard-monitoring shipyard-observability
```

The three Shipyard services should expose metrics internally:

```powershell
kubectl -n shipyard port-forward svc/shipyard-auth-service 8081:8081
curl.exe http://127.0.0.1:8081/metrics
```

Repeat the same check with:

- `svc/shipyard-todo-service` on port `8080`
- `svc/shipyard-platform-status` on port `8082`

## Open Grafana

Port-forward Grafana:

```powershell
kubectl -n monitoring port-forward svc/shipyard-monitoring-grafana 3001:80
```

Open:

```text
http://127.0.0.1:3001
```

Get the generated Grafana credentials from the in-cluster secret:

```powershell
$user = kubectl -n monitoring get secret shipyard-monitoring-grafana -o jsonpath="{.data.admin-user}"
$password = kubectl -n monitoring get secret shipyard-monitoring-grafana -o jsonpath="{.data.admin-password}"
[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($user))
[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($password))
```

Use the `Shipyard HTTP Overview` dashboard for portfolio screenshots. Default
Kubernetes dashboards from `kube-prometheus-stack` can be used to show node,
pod, CPU, and memory evidence.

## Teardown

When monitoring is not needed, remove the Argo CD Applications. The manifests
include Argo CD resource finalizers so the managed resources are pruned with the
Applications.

```powershell
kubectl -n argocd delete application shipyard-observability
kubectl -n argocd delete application shipyard-monitoring
doctl kubernetes cluster node-pool update shipyard-doks-demo demo-pool --count 1 --min-nodes 1 --max-nodes 1 --auto-scale
```

This keeps the live demo cost-conscious and avoids leaving Prometheus/Grafana
running when they are not needed.

## Do Not Publish

Do not publish screenshots or logs that show:

- Grafana admin password
- Kubernetes secrets
- service account tokens
- kubeconfig
- database credentials
- GitHub or DigitalOcean tokens
