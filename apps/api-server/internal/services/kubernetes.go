package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/circuitbreaker"
)

// KubernetesService provides real Kubernetes cluster operations using client-go
type KubernetesService struct {
	clientset      *kubernetes.Clientset
	namespace      string
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// NewKubernetesService creates a Kubernetes client using in-cluster config or kubeconfig
func NewKubernetesService() (*KubernetesService, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	if config, err = rest.InClusterConfig(); err != nil {
		// Fall back to kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	// Determine namespace
	namespace := os.Getenv("KUBERNETES_NAMESPACE")
	if namespace == "" {
		if ns, err := getInClusterNamespace(); err == nil {
			namespace = ns
		} else {
			namespace = "default"
		}
	}

	// Initialize circuit breaker for Kubernetes API calls
	cbConfig := circuitbreaker.Config{
		FailureThreshold: getEnvUint("K8S_CB_FAILURE_THRESHOLD", 5),
		SuccessThreshold: getEnvUint("K8S_CB_SUCCESS_THRESHOLD", 2),
		Timeout:          time.Duration(getEnvUint("K8S_CB_TIMEOUT_SECONDS", 60)) * time.Second,
	}
	circuitBreaker := circuitbreaker.NewCircuitBreaker(cbConfig)

	return &KubernetesService{
		clientset:      clientset,
		namespace:      namespace,
		circuitBreaker: circuitBreaker,
	}, nil
}

func getEnvUint(key string, defaultVal uint) uint {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var result uint
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		return defaultVal
	}
	return result
}

// getInClusterNamespace reads the namespace from the in-cluster service account
func getInClusterNamespace() (string, error) {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Namespace returns the configured namespace
func (k *KubernetesService) Namespace() string {
	return k.namespace
}

// Deployment wraps Kubernetes Deployment fields we expose via API
type Deployment struct {
	Name                string            `json:"name"`
	Namespace           string            `json:"namespace"`
	CreationTimestamp   time.Time         `json:"creationTimestamp"`
	Labels              map[string]string `json:"labels"`
	Replicas            int32             `json:"replicas"`
	ReadyReplicas       int32             `json:"readyReplicas"`
	AvailableReplicas   int32             `json:"availableReplicas"`
	UpdatedReplicas     int32             `json:"updatedReplicas"`
	UnavailableReplicas int32             `json:"unavailableReplicas"`
}

// ListDeployments returns all deployments in the configured namespace
func (k *KubernetesService) ListDeployments(ctx context.Context) ([]Deployment, error) {
	deployments, err := k.clientset.AppsV1().Deployments(k.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var result []Deployment
	for _, d := range deployments.Items {
		result = append(result, Deployment{
			Name:                d.Name,
			Namespace:           d.Namespace,
			CreationTimestamp:   d.CreationTimestamp.Time,
			Labels:              d.Labels,
			Replicas:            *d.Spec.Replicas,
			ReadyReplicas:       d.Status.ReadyReplicas,
			AvailableReplicas:   d.Status.AvailableReplicas,
			UpdatedReplicas:     d.Status.UpdatedReplicas,
			UnavailableReplicas: d.Status.UnavailableReplicas,
		})
	}

	return result, nil
}

// GetDeployment returns a specific deployment by name
func (k *KubernetesService) GetDeployment(ctx context.Context, name string) (*Deployment, error) {
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", name, err)
	}

	return &Deployment{
		Name:                deployment.Name,
		Namespace:           deployment.Namespace,
		CreationTimestamp:   deployment.CreationTimestamp.Time,
		Labels:              deployment.Labels,
		Replicas:            *deployment.Spec.Replicas,
		ReadyReplicas:       deployment.Status.ReadyReplicas,
		AvailableReplicas:   deployment.Status.AvailableReplicas,
		UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		UnavailableReplicas: deployment.Status.UnavailableReplicas,
	}, nil
}

// ScaleDeployment updates the replica count for a deployment
func (k *KubernetesService) ScaleDeployment(ctx context.Context, name string, replicas int32) error {
	if replicas < 0 || replicas > 100 {
		return fmt.Errorf("replica count %d is out of valid range [0, 100]", replicas)
	}

	// Get current deployment
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s for scaling: %w", name, err)
	}

	// Update replica count
	deployment.Spec.Replicas = &replicas
	_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment %s to %d replicas: %w", name, replicas, err)
	}

	return nil
}

// RestartDeployment triggers a rolling restart by updating the pod template annotation
func (k *KubernetesService) RestartDeployment(ctx context.Context, name string) error {
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s for restart: %w", name, err)
	}

	// Add restart annotation to trigger rolling restart
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)

	_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart deployment %s: %w", name, err)
	}

	return nil
}

// PodLogs fetches recent logs from pods matching the deployment's label selector
func (k *KubernetesService) PodLogs(ctx context.Context, name string, tailLines int) (string, error) {
	// Get deployment to find label selector
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %s for logs: %w", name, err)
	}

	// Find pods using deployment's label selector
	labelSelector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector()
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods for deployment %s: %w", name, err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for deployment %s", name)
	}

	// Get logs from the first running pod
	var pod *corev1.Pod
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodRunning {
			pod = &p
			break
		}
	}
	if pod == nil {
		// If no running pod, use the first one
		pod = &pods.Items[0]
	}

	if tailLines <= 0 {
		tailLines = 100
	}

	tailLines64 := int64(tailLines)
	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		TailLines: &tailLines64,
		Follow:    false,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s: %w", pod.Name, err)
	}
	defer logs.Close()

	data := make([]byte, 1024*1024) // 1MB limit
	n, err := logs.Read(data)
	if err != nil && err.Error() != "EOF" {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(data[:n]), nil
}

// GetPodStatus returns detailed pod status for a deployment
func (k *KubernetesService) GetPodStatus(ctx context.Context, name string) (map[string]any, error) {
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", name, err)
	}

	labelSelector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector()
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for deployment %s: %w", name, err)
	}

	status := map[string]any{
		"deployment": map[string]any{
			"replicas":             *deployment.Spec.Replicas,
			"ready_replicas":       deployment.Status.ReadyReplicas,
			"available_replicas":   deployment.Status.AvailableReplicas,
			"updated_replicas":     deployment.Status.UpdatedReplicas,
			"unavailable_replicas": deployment.Status.UnavailableReplicas,
		},
		"pods": make([]map[string]any, 0),
	}

	for _, pod := range pods.Items {
		podStatus := map[string]any{
			"name":       pod.Name,
			"phase":      string(pod.Status.Phase),
			"node_name":  pod.Spec.NodeName,
			"created":    pod.CreationTimestamp.Time,
			"conditions": make([]map[string]any, 0),
		}

		for _, condition := range pod.Status.Conditions {
			podStatus["conditions"] = append(podStatus["conditions"].([]map[string]any), map[string]any{
				"type":   string(condition.Type),
				"status": string(condition.Status),
				"reason": condition.Reason,
			})
		}

		status["pods"] = append(status["pods"].([]map[string]any), podStatus)
	}

	return status, nil
}

// GetEvents returns recent events for a deployment
func (k *KubernetesService) GetEvents(ctx context.Context, name string, limit int) ([]map[string]any, error) {
	events, err := k.clientset.CoreV1().Events(k.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", name),
		Limit:         int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for deployment %s: %w", name, err)
	}

	result := make([]map[string]any, 0)
	for _, event := range events.Items {
		result = append(result, map[string]any{
			"type":            event.Type,
			"reason":          event.Reason,
			"message":         event.Message,
			"first_timestamp": event.FirstTimestamp.Time,
			"last_timestamp":  event.LastTimestamp.Time,
			"count":           event.Count,
		})
	}

	return result, nil
}
