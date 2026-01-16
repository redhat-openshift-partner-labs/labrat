// Package hub provides functionality for interacting with Red Hat Advanced Cluster Management (ACM)
// ManagedCluster resources. It includes types, client operations, status derivation logic,
// and output formatting for managed cluster information.
package hub

// ClusterStatus represents the overall status of a managed cluster
type ClusterStatus string

const (
	// StatusReady indicates the cluster is available and ready
	StatusReady ClusterStatus = "Ready"
	// StatusNotReady indicates the cluster is not available or has issues
	StatusNotReady ClusterStatus = "NotReady"
	// StatusUnknown indicates the cluster status cannot be determined
	StatusUnknown ClusterStatus = "Unknown"
)

// ManagedClusterInfo contains information about a managed cluster
type ManagedClusterInfo struct {
	// Name is the name of the managed cluster
	Name string
	// Status is the overall status of the cluster
	Status ClusterStatus
	// Available is the value of the ManagedClusterConditionAvailable condition
	Available string
	// Message provides additional context about the cluster status
	Message string
}

// ManagedClusterFilter defines criteria for filtering managed clusters
type ManagedClusterFilter struct {
	// Status filters clusters by their overall status
	Status ClusterStatus
}

// ClusterDeploymentInfo contains information from a Hive ClusterDeployment resource
type ClusterDeploymentInfo struct {
	// Name is the name of the cluster deployment
	Name string
	// Namespace is the namespace containing the cluster deployment (typically same as Name)
	Namespace string
	// PowerState indicates if the cluster is running or hibernating
	PowerState string
	// Installed indicates whether the cluster installation is complete
	Installed bool
	// APIUrl is the Kubernetes API server URL
	APIUrl string
	// ConsoleURL is the OpenShift console URL
	ConsoleURL string
	// KubeconfigSecretName is the name of the secret containing the admin kubeconfig
	KubeconfigSecretName string
	// KubeconfigSecretNS is the namespace of the kubeconfig secret
	KubeconfigSecretNS string
	// Platform is the cloud platform (AWS, Azure, GCP, etc.)
	Platform string
	// Region is the cloud region
	Region string
	// Version is the OpenShift version
	Version string
}

// CombinedClusterInfo merges information from both ManagedCluster and ClusterDeployment
type CombinedClusterInfo struct {
	// Name is the cluster name
	Name string
	// Status is the overall cluster status from ManagedCluster
	Status ClusterStatus
	// PowerState indicates if the cluster is running or hibernating from ClusterDeployment
	PowerState string
	// Platform is the cloud platform from ClusterDeployment
	Platform string
	// Region is the cloud region from ClusterDeployment
	Region string
	// Version is the OpenShift version from ClusterDeployment
	Version string
	// APIUrl is the Kubernetes API server URL from ClusterDeployment
	APIUrl string
	// ConsoleURL is the OpenShift console URL from ClusterDeployment
	ConsoleURL string
	// Available is the ManagedClusterConditionAvailable status from ManagedCluster
	Available string
	// KubeconfigSecret is the namespace/name of the admin kubeconfig secret
	KubeconfigSecret string
	// Message provides additional context about the cluster status
	Message string
}
