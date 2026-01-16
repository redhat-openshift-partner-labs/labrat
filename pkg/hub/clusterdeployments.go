package hub

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ClusterDeploymentClient provides operations for interacting with Hive ClusterDeployment resources
type ClusterDeploymentClient interface {
	// Get retrieves a ClusterDeployment by name from the namespace with the same name
	Get(ctx context.Context, name string) (*ClusterDeploymentInfo, error)
}

type clusterDeploymentClient struct {
	dynamicClient dynamic.Interface
}

// NewClusterDeploymentClient creates a new ClusterDeploymentClient
func NewClusterDeploymentClient(dynamicClient dynamic.Interface) ClusterDeploymentClient {
	return &clusterDeploymentClient{
		dynamicClient: dynamicClient,
	}
}

// Get retrieves a ClusterDeployment from the namespace matching the cluster name
func (c *clusterDeploymentClient) Get(ctx context.Context, name string) (*ClusterDeploymentInfo, error) {
	// Define the GVR for ClusterDeployment
	gvr := schema.GroupVersionResource{
		Group:    "hive.openshift.io",
		Version:  "v1",
		Resource: "clusterdeployments",
	}

	// Get the ClusterDeployment from namespace=name
	unstructuredCD, err := c.dynamicClient.Resource(gvr).Namespace(name).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterDeployment %s: %w", name, err)
	}

	// Parse the unstructured object into ClusterDeploymentInfo
	info, err := parseClusterDeployment(unstructuredCD.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ClusterDeployment %s: %w", name, err)
	}

	return info, nil
}

// parseClusterDeployment extracts ClusterDeploymentInfo from an unstructured object
func parseClusterDeployment(obj map[string]interface{}) (*ClusterDeploymentInfo, error) {
	info := &ClusterDeploymentInfo{}

	// Extract metadata
	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("metadata not found or invalid")
	}

	if name, ok := metadata["name"].(string); ok {
		info.Name = name
	}

	if namespace, ok := metadata["namespace"].(string); ok {
		info.Namespace = namespace
	}

	// Extract labels for platform and region
	if labels, ok := metadata["labels"].(map[string]interface{}); ok {
		if platform, ok := labels["hive.openshift.io/cluster-platform"].(string); ok {
			info.Platform = platform
		}
		if region, ok := labels["hive.openshift.io/cluster-region"].(string); ok {
			info.Region = region
		}
	}

	// Extract spec fields
	if spec, ok := obj["spec"].(map[string]interface{}); ok {
		// Power state from spec
		if powerState, ok := spec["powerState"].(string); ok {
			info.PowerState = powerState
		}

		// Installed status
		if installed, ok := spec["installed"].(bool); ok {
			info.Installed = installed
		}

		// Extract kubeconfig secret reference from clusterMetadata
		if clusterMetadata, ok := spec["clusterMetadata"].(map[string]interface{}); ok {
			if adminKubeconfigRef, ok := clusterMetadata["adminKubeconfigSecretRef"].(map[string]interface{}); ok {
				if name, ok := adminKubeconfigRef["name"].(string); ok {
					info.KubeconfigSecretName = name
					// Secret is in the same namespace as the ClusterDeployment
					info.KubeconfigSecretNS = info.Namespace
				}
			}
		}
	}

	// Extract status fields
	if status, ok := obj["status"].(map[string]interface{}); ok {
		if apiURL, ok := status["apiURL"].(string); ok {
			info.APIUrl = apiURL
		}

		if consoleURL, ok := status["webConsoleURL"].(string); ok {
			info.ConsoleURL = consoleURL
		}

		if version, ok := status["installVersion"].(string); ok {
			info.Version = version
		}

		// Power state from status (takes precedence over spec)
		if powerState, ok := status["powerState"].(string); ok {
			info.PowerState = powerState
		}
	}

	// Default power state if not specified
	if info.PowerState == "" {
		info.PowerState = "Unknown"
	}

	return info, nil
}
