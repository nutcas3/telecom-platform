package kubernetes

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// getInClusterNamespace reads the namespace from the in-cluster service account
func getInClusterNamespace() (string, error) {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListServices lists all services/deployments in the namespace
func (sm *ServiceManager) ListServices() ([]Service, error) {
	deployments, err := sm.clientset.AppsV1().Deployments(sm.namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	services := make([]Service, 0, len(deployments.Items))
	now := time.Now()

	for _, d := range deployments.Items {
		uptime := "0m"
		if !d.CreationTimestamp.IsZero() {
			uptime = now.Sub(d.CreationTimestamp.Time).Round(time.Minute).String()
		}

		status := "Running"
		if d.Status.UnavailableReplicas > 0 {
			status = "Degraded"
		}
		if d.Spec.Replicas != nil && *d.Spec.Replicas == 0 {
			status = "Stopped"
		}

		version := "v1.0.0"
		if d.Annotations != nil {
			if v, ok := d.Annotations["version"]; ok {
				version = v
			}
		}

		services = append(services, Service{
			Name:      d.Name,
			Status:    status,
			Version:   version,
			Uptime:    uptime,
			Replicas:  *d.Spec.Replicas,
			Available: d.Status.AvailableReplicas,
		})
	}

	return services, nil
}

// GetServiceStatus gets detailed status for a specific service
func (sm *ServiceManager) GetServiceStatus(serviceName string) (*Service, error) {
	deployment, err := sm.clientset.AppsV1().Deployments(sm.namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", serviceName, err)
	}

	now := time.Now()
	uptime := "0m"
	if !deployment.CreationTimestamp.IsZero() {
		uptime = now.Sub(deployment.CreationTimestamp.Time).Round(time.Minute).String()
	}

	status := "Running"
	if deployment.Status.UnavailableReplicas > 0 {
		status = "Degraded"
	}
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
		status = "Stopped"
	}

	version := "v1.0.0"
	if deployment.Annotations != nil {
		if v, ok := deployment.Annotations["version"]; ok {
			version = v
		}
	}

	// Get pods for CPU/memory info
	pods, err := sm.clientset.CoreV1().Pods(sm.namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector().String(),
	})
	if err == nil && len(pods.Items) > 0 {
		// Simple estimation - in production use metrics API
		cpu := 45.0
		memory := "256MB"
		return &Service{
			Name:      deployment.Name,
			Status:    status,
			Version:   version,
			Uptime:    uptime,
			CPU:       cpu,
			Memory:    memory,
			Replicas:  *deployment.Spec.Replicas,
			Available: deployment.Status.AvailableReplicas,
		}, nil
	}

	return &Service{
		Name:      deployment.Name,
		Status:    status,
		Version:   version,
		Uptime:    uptime,
		CPU:       0.0,
		Memory:    "0MB",
		Replicas:  *deployment.Spec.Replicas,
		Available: deployment.Status.AvailableReplicas,
	}, nil
}

// RestartDeployment restarts a deployment by rolling its pods
func (sm *ServiceManager) RestartDeployment(serviceName string) error {
	deployment, err := sm.clientset.AppsV1().Deployments(sm.namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", serviceName, err)
	}

	// Add restart annotation to trigger rolling restart
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)

	_, err = sm.clientset.AppsV1().Deployments(sm.namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart deployment %s: %w", serviceName, err)
	}

	return nil
}

// GetPodLogs fetches recent logs from pods for a deployment
func (sm *ServiceManager) GetPodLogs(serviceName string, tailLines int64) ([]string, error) {
	deployment, err := sm.clientset.AppsV1().Deployments(sm.namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", serviceName, err)
	}

	labelSelector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector()
	pods, err := sm.clientset.CoreV1().Pods(sm.namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for deployment %s: %w", serviceName, err)
	}

	if len(pods.Items) == 0 {
		return []string{"No pods found for deployment " + serviceName}, nil
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
		pod = &pods.Items[0] // Use first pod even if not running
	}

	req := sm.clientset.CoreV1().Pods(sm.namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		TailLines: &tailLines,
	})

	logData, err := req.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs from pod %s: %w", pod.Name, err)
	}
	defer logData.Close()

	buf := make([]byte, 1024*1024) // 1MB limit
	n, err := logData.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	logs := string(buf[:n])
	// Split logs into lines
	logLines := []string{}
	currentLine := ""
	for _, c := range logs {
		if c == '\n' {
			logLines = append(logLines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(c)
		}
	}
	if currentLine != "" {
		logLines = append(logLines, currentLine)
	}

	// Return last 20 lines if there are more
	if len(logLines) > 20 {
		logLines = logLines[len(logLines)-20:]
	}

	return logLines, nil
}
