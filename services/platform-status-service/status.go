package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultTokenFile     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultCAFile        = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	defaultNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type SnapshotProvider interface {
	Snapshot(ctx context.Context) (PlatformSnapshot, error)
}

type PlatformSnapshot struct {
	CheckedAt time.Time         `json:"checkedAt"`
	Cluster   ClusterSummary    `json:"cluster"`
	Namespace string            `json:"namespace"`
	Workloads []WorkloadSummary `json:"workloads"`
	Services  []ServiceSummary  `json:"services"`
	Storage   []StorageSummary  `json:"storage"`
	GitOps    []GitOpsSummary   `json:"gitops"`
	Warnings  []string          `json:"warnings,omitempty"`
}

type ClusterSummary struct {
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Mode     string `json:"mode"`
}

type WorkloadSummary struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Ready  string `json:"ready"`
	Status string `json:"status"`
}

type ServiceSummary struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Ports []string `json:"ports"`
}

type StorageSummary struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Size   string `json:"size"`
}

type GitOpsSummary struct {
	Name   string `json:"name"`
	Sync   string `json:"sync"`
	Health string `json:"health"`
}

type SnapshotConfig struct {
	Namespace        string
	ArgoCDNamespace  string
	ClusterProvider  string
	ClusterRegion    string
	ClusterMode      string
	ExpectedServices []string
}

func SnapshotConfigFromEnv() SnapshotConfig {
	ns := firstNonEmpty(os.Getenv("SHIPYARD_NAMESPACE"), readNamespaceFile(), "shipyard")
	return SnapshotConfig{
		Namespace:       ns,
		ArgoCDNamespace: firstNonEmpty(os.Getenv("ARGOCD_NAMESPACE"), "argocd"),
		ClusterProvider: firstNonEmpty(os.Getenv("CLUSTER_PROVIDER"), "DigitalOcean Kubernetes"),
		ClusterRegion:   firstNonEmpty(os.Getenv("CLUSTER_REGION"), "sgp1"),
		ClusterMode:     firstNonEmpty(os.Getenv("CLUSTER_MODE"), "cost-conscious demo"),
		ExpectedServices: []string{
			"shipyard-frontend",
			"shipyard-auth-service",
			"shipyard-todo-service",
			"postgres",
		},
	}
}

func PlatformStatusHandler(provider SnapshotProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		snapshot, err := provider.Snapshot(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "live infrastructure snapshot unavailable",
			})
			return
		}
		c.JSON(http.StatusOK, snapshot)
	}
}

type unavailableProvider struct{}

func (unavailableProvider) Snapshot(context.Context) (PlatformSnapshot, error) {
	return PlatformSnapshot{}, errors.New("snapshot provider unavailable")
}

type KubeClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewInClusterClient() (*KubeClient, error) {
	baseURL := strings.TrimRight(os.Getenv("KUBE_API_BASE_URL"), "/")
	if baseURL == "" {
		host := os.Getenv("KUBERNETES_SERVICE_HOST")
		port := firstNonEmpty(os.Getenv("KUBERNETES_SERVICE_PORT"), "443")
		if host == "" {
			return nil, errors.New("kubernetes service host is not set")
		}
		baseURL = fmt.Sprintf("https://%s:%s", host, port)
	}

	token := os.Getenv("KUBE_BEARER_TOKEN")
	if token == "" {
		tokenFile := firstNonEmpty(os.Getenv("KUBE_TOKEN_FILE"), defaultTokenFile)
		raw, err := os.ReadFile(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("read service account token: %w", err)
		}
		token = strings.TrimSpace(string(raw))
	}
	if token == "" {
		return nil, errors.New("service account token is empty")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	caFile := firstNonEmpty(os.Getenv("KUBE_CA_FILE"), defaultCAFile)
	if caPEM, err := os.ReadFile(caFile); err == nil {
		roots := x509.NewCertPool()
		if roots.AppendCertsFromPEM(caPEM) {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    roots,
			}
		}
	}

	return &KubeClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		},
	}, nil
}

func (c *KubeClient) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("kubernetes api returned status %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(out)
}

type KubeSnapshotProvider struct {
	client *KubeClient
	config SnapshotConfig
}

func NewKubeSnapshotProvider(client *KubeClient, config SnapshotConfig) KubeSnapshotProvider {
	if len(config.ExpectedServices) == 0 {
		config.ExpectedServices = SnapshotConfigFromEnv().ExpectedServices
	}
	return KubeSnapshotProvider{client: client, config: config}
}

func (p KubeSnapshotProvider) Snapshot(ctx context.Context) (PlatformSnapshot, error) {
	snapshot := PlatformSnapshot{
		CheckedAt: time.Now().UTC(),
		Cluster: ClusterSummary{
			Provider: p.config.ClusterProvider,
			Region:   p.config.ClusterRegion,
			Mode:     p.config.ClusterMode,
		},
		Namespace: p.config.Namespace,
	}

	if err := p.addDeployments(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "deployments unavailable")
	}
	if err := p.addStatefulSets(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "statefulsets unavailable")
	}
	if err := p.addRollouts(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "rollouts unavailable")
	}
	if err := p.addServices(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "services unavailable")
	}
	if err := p.addStorage(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "storage unavailable")
	}
	if err := p.addGitOps(ctx, &snapshot); err != nil {
		snapshot.Warnings = append(snapshot.Warnings, "gitops status unavailable")
	}

	sort.Slice(snapshot.Workloads, func(i, j int) bool {
		return workloadOrder(snapshot.Workloads[i]) < workloadOrder(snapshot.Workloads[j])
	})
	sort.Slice(snapshot.Services, func(i, j int) bool {
		return serviceOrder(snapshot.Services[i].Name) < serviceOrder(snapshot.Services[j].Name)
	})
	sort.Slice(snapshot.Storage, func(i, j int) bool {
		return snapshot.Storage[i].Name < snapshot.Storage[j].Name
	})
	sort.Slice(snapshot.GitOps, func(i, j int) bool {
		return snapshot.GitOps[i].Name < snapshot.GitOps[j].Name
	})

	if len(snapshot.Workloads) == 0 && len(snapshot.Services) == 0 && len(snapshot.Storage) == 0 && len(snapshot.GitOps) == 0 {
		return PlatformSnapshot{}, errors.New("no platform resources available")
	}

	return snapshot, nil
}

func (p KubeSnapshotProvider) addDeployments(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list deploymentList
	if err := p.client.getJSON(ctx, "/apis/apps/v1/namespaces/"+p.config.Namespace+"/deployments", &list); err != nil {
		return err
	}

	for _, item := range list.Items {
		switch item.Metadata.Name {
		case "shipyard-frontend", "shipyard-auth-service":
			desired := desiredReplicas(item.Spec.Replicas, item.Status.Replicas)
			snapshot.Workloads = append(snapshot.Workloads, WorkloadSummary{
				Name:   item.Metadata.Name,
				Kind:   "Deployment",
				Ready:  readyString(item.Status.ReadyReplicas, desired),
				Status: healthStatus(item.Status.ReadyReplicas, desired),
			})
		}
	}
	return nil
}

func (p KubeSnapshotProvider) addStatefulSets(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list statefulSetList
	if err := p.client.getJSON(ctx, "/apis/apps/v1/namespaces/"+p.config.Namespace+"/statefulsets", &list); err != nil {
		return err
	}

	for _, item := range list.Items {
		if item.Metadata.Name != "postgres" {
			continue
		}
		desired := desiredReplicas(item.Spec.Replicas, item.Status.Replicas)
		snapshot.Workloads = append(snapshot.Workloads, WorkloadSummary{
			Name:   item.Metadata.Name,
			Kind:   "StatefulSet",
			Ready:  readyString(item.Status.ReadyReplicas, desired),
			Status: healthStatus(item.Status.ReadyReplicas, desired),
		})
	}
	return nil
}

func (p KubeSnapshotProvider) addRollouts(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list rolloutList
	if err := p.client.getJSON(ctx, "/apis/argoproj.io/v1alpha1/namespaces/"+p.config.Namespace+"/rollouts", &list); err != nil {
		return err
	}

	for _, item := range list.Items {
		if item.Metadata.Name != "shipyard-todo-service" {
			continue
		}
		desired := desiredReplicas(item.Spec.Replicas, item.Status.Replicas)
		status := healthStatus(item.Status.ReadyReplicas, desired)
		if item.Status.Phase != "" && status != "Healthy" {
			status = item.Status.Phase
		}
		snapshot.Workloads = append(snapshot.Workloads, WorkloadSummary{
			Name:   item.Metadata.Name,
			Kind:   "Rollout",
			Ready:  readyString(item.Status.ReadyReplicas, desired),
			Status: status,
		})
	}
	return nil
}

func (p KubeSnapshotProvider) addServices(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list serviceList
	if err := p.client.getJSON(ctx, "/api/v1/namespaces/"+p.config.Namespace+"/services", &list); err != nil {
		return err
	}

	expected := toSet(p.config.ExpectedServices)
	for _, item := range list.Items {
		if !expected[item.Metadata.Name] {
			continue
		}
		ports := make([]string, 0, len(item.Spec.Ports))
		for _, port := range item.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d", port.Port))
		}
		snapshot.Services = append(snapshot.Services, ServiceSummary{
			Name:  item.Metadata.Name,
			Type:  firstNonEmpty(item.Spec.Type, "ClusterIP"),
			Ports: ports,
		})
	}
	return nil
}

func (p KubeSnapshotProvider) addStorage(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list pvcList
	if err := p.client.getJSON(ctx, "/api/v1/namespaces/"+p.config.Namespace+"/persistentvolumeclaims", &list); err != nil {
		return err
	}

	for _, item := range list.Items {
		if item.Metadata.Name != "data-postgres-0" {
			continue
		}
		snapshot.Storage = append(snapshot.Storage, StorageSummary{
			Name:   item.Metadata.Name,
			Status: firstNonEmpty(item.Status.Phase, "Unknown"),
			Size:   firstNonEmpty(item.Spec.Resources.Requests["storage"], "unknown"),
		})
	}
	return nil
}

func (p KubeSnapshotProvider) addGitOps(ctx context.Context, snapshot *PlatformSnapshot) error {
	var list applicationList
	if err := p.client.getJSON(ctx, "/apis/argoproj.io/v1alpha1/namespaces/"+p.config.ArgoCDNamespace+"/applications", &list); err != nil {
		return err
	}

	for _, item := range list.Items {
		if item.Metadata.Name != "argo-rollouts" && !strings.HasPrefix(item.Metadata.Name, "shipyard-") {
			continue
		}
		snapshot.GitOps = append(snapshot.GitOps, GitOpsSummary{
			Name:   item.Metadata.Name,
			Sync:   firstNonEmpty(item.Status.Sync.Status, "Unknown"),
			Health: firstNonEmpty(item.Status.Health.Status, "Unknown"),
		})
	}
	return nil
}

type objectMeta struct {
	Name string `json:"name"`
}

type replicaSpec struct {
	Replicas *int32 `json:"replicas"`
}

type workloadStatus struct {
	ReadyReplicas int32  `json:"readyReplicas"`
	Replicas      int32  `json:"replicas"`
	Phase         string `json:"phase"`
}

type deploymentList struct {
	Items []struct {
		Metadata objectMeta     `json:"metadata"`
		Spec     replicaSpec    `json:"spec"`
		Status   workloadStatus `json:"status"`
	} `json:"items"`
}

type statefulSetList struct {
	Items []struct {
		Metadata objectMeta     `json:"metadata"`
		Spec     replicaSpec    `json:"spec"`
		Status   workloadStatus `json:"status"`
	} `json:"items"`
}

type rolloutList struct {
	Items []struct {
		Metadata objectMeta     `json:"metadata"`
		Spec     replicaSpec    `json:"spec"`
		Status   workloadStatus `json:"status"`
	} `json:"items"`
}

type serviceList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			Type  string `json:"type"`
			Ports []struct {
				Port int32 `json:"port"`
			} `json:"ports"`
		} `json:"spec"`
	} `json:"items"`
}

type pvcList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			Resources struct {
				Requests map[string]string `json:"requests"`
			} `json:"resources"`
		} `json:"spec"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}

type applicationList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Status   struct {
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"status"`
	} `json:"items"`
}

func desiredReplicas(spec *int32, status int32) int32 {
	if spec != nil {
		return *spec
	}
	if status > 0 {
		return status
	}
	return 1
}

func readyString(ready, desired int32) string {
	return fmt.Sprintf("%d/%d", ready, desired)
}

func healthStatus(ready, desired int32) string {
	if desired == 0 {
		return "ScaledDown"
	}
	if ready >= desired {
		return "Healthy"
	}
	if ready > 0 {
		return "Progressing"
	}
	return "Unavailable"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func readNamespaceFile() string {
	raw, err := os.ReadFile(defaultNamespaceFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func workloadOrder(workload WorkloadSummary) int {
	switch workload.Name {
	case "shipyard-frontend":
		return 10
	case "shipyard-auth-service":
		return 20
	case "shipyard-todo-service":
		return 30
	case "postgres":
		return 40
	default:
		return 100
	}
}

func serviceOrder(name string) int {
	switch name {
	case "shipyard-frontend":
		return 10
	case "shipyard-auth-service":
		return 20
	case "shipyard-todo-service":
		return 30
	case "postgres":
		return 40
	default:
		return 100
	}
}
