package kubernetes

import "k8s.io/client-go/kubernetes"

// ServiceManager handles Kubernetes service operations
type ServiceManager struct {
	clientset  *kubernetes.Clientset
	namespace string
}

// Service represents a Kubernetes service/deployment status
type Service struct {
	Name      string
	Status    string
	Version   string
	Uptime    string
	CPU       float64
	Memory    string
	Replicas  int32
	Available int32
}