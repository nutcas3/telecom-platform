package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewServiceManager creates a new Kubernetes service manager
func NewServiceManager() (*ServiceManager, error) {
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
			namespace = "telecom-platform-dev"
		}
	}

	return &ServiceManager{
		clientset:  clientset,
		namespace: namespace,
	}, nil
}

// ScaleDeployment scales a deployment to the specified replica count
func (sm *ServiceManager) ScaleDeployment(serviceName string, replicas int32) error {
	deployment, err := sm.clientset.AppsV1().Deployments(sm.namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", serviceName, err)
	}

	deployment.Spec.Replicas = &replicas
	_, err = sm.clientset.AppsV1().Deployments(sm.namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment %s: %w", serviceName, err)
	}

	return nil
}

// StartAllServices scales all platform services up
func (sm *ServiceManager) StartAllServices() error {
	services := []string{"api-server", "charging-engine", "carrier-connector", "packet-gateway", "web-dashboard"}
	replicas := int32(1)

	for _, svc := range services {
		err := sm.ScaleDeployment(svc, replicas)
		if err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc, err)
		}
	}

	return nil
}

// StopAllServices scales all platform services down
func (sm *ServiceManager) StopAllServices() error {
	services := []string{"web-dashboard", "packet-gateway", "carrier-connector", "charging-engine", "api-server"}
	replicas := int32(0)

	for _, svc := range services {
		err := sm.ScaleDeployment(svc, replicas)
		if err != nil {
			return fmt.Errorf("failed to stop service %s: %w", svc, err)
		}
	}

	return nil
}
