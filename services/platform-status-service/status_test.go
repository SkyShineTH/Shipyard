package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestKubeSnapshotProviderReturnsSanitizedSnapshot(t *testing.T) {
	t.Parallel()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis/apps/v1/namespaces/shipyard/deployments":
			writeJSON(t, w, map[string]any{"items": []any{
				workload("shipyard-frontend", 1, 1),
				workload("shipyard-auth-service", 1, 1),
				workload("internal-debug", 1, 1),
			}})
		case "/apis/apps/v1/namespaces/shipyard/statefulsets":
			writeJSON(t, w, map[string]any{"items": []any{workload("postgres", 1, 1)}})
		case "/apis/argoproj.io/v1alpha1/namespaces/shipyard/rollouts":
			writeJSON(t, w, map[string]any{"items": []any{workload("shipyard-todo-service", 1, 1)}})
		case "/api/v1/namespaces/shipyard/services":
			writeJSON(t, w, map[string]any{"items": []any{
				map[string]any{
					"metadata": map[string]any{"name": "shipyard-frontend"},
					"spec": map[string]any{
						"type":      "LoadBalancer",
						"clusterIP": "10.245.14.20",
						"ports": []any{
							map[string]any{"port": 80},
							map[string]any{"port": 443},
						},
					},
				},
				map[string]any{
					"metadata": map[string]any{"name": "kubernetes"},
					"spec": map[string]any{
						"type": "ClusterIP",
						"ports": []any{
							map[string]any{"port": 443},
						},
					},
				},
			}})
		case "/api/v1/namespaces/shipyard/persistentvolumeclaims":
			writeJSON(t, w, map[string]any{"items": []any{
				map[string]any{
					"metadata": map[string]any{"name": "data-postgres-0"},
					"spec": map[string]any{
						"resources": map[string]any{
							"requests": map[string]any{"storage": "2Gi"},
						},
					},
					"status": map[string]any{"phase": "Bound"},
				},
			}})
		case "/apis/argoproj.io/v1alpha1/namespaces/argocd/applications":
			writeJSON(t, w, map[string]any{"items": []any{
				map[string]any{
					"metadata": map[string]any{"name": "shipyard-frontend"},
					"status": map[string]any{
						"sync":   map[string]any{"status": "Synced"},
						"health": map[string]any{"status": "Healthy"},
					},
				},
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(api.Close)

	provider := NewKubeSnapshotProvider(&KubeClient{
		baseURL:    api.URL,
		token:      "test-token",
		httpClient: api.Client(),
	}, SnapshotConfig{
		Namespace:       "shipyard",
		ArgoCDNamespace: "argocd",
		ClusterProvider: "DigitalOcean Kubernetes",
		ClusterRegion:   "sgp1",
		ClusterMode:     "cost-conscious demo",
		ExpectedServices: []string{
			"shipyard-frontend",
			"shipyard-auth-service",
			"shipyard-todo-service",
			"postgres",
		},
	})

	snapshot, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}

	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	body := string(raw)
	for _, forbidden := range []string{"10.245.14.20", "internal-debug", "kubernetes"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("snapshot leaked forbidden value %q: %s", forbidden, body)
		}
	}

	if len(snapshot.Workloads) != 4 {
		t.Fatalf("expected 4 workloads, got %d", len(snapshot.Workloads))
	}
	if len(snapshot.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(snapshot.Services))
	}
	if snapshot.Services[0].Name != "shipyard-frontend" {
		t.Fatalf("unexpected service: %s", snapshot.Services[0].Name)
	}
	if strings.Join(snapshot.Services[0].Ports, ",") != "80,443" {
		t.Fatalf("unexpected service ports: %v", snapshot.Services[0].Ports)
	}
}

func TestPlatformStatusHandlerReturnsGenericUnavailableError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/platform/status", PlatformStatusHandler(errorProvider{}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/platform/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "live infrastructure snapshot unavailable") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

type errorProvider struct{}

func (errorProvider) Snapshot(context.Context) (PlatformSnapshot, error) {
	return PlatformSnapshot{CheckedAt: time.Now().UTC()}, errors.New("token file missing at /secret/path")
}

func workload(name string, desired int, ready int) map[string]any {
	return map[string]any{
		"metadata": map[string]any{"name": name},
		"spec":     map[string]any{"replicas": desired},
		"status": map[string]any{
			"readyReplicas": ready,
			"replicas":      desired,
		},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}
