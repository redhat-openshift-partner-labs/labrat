package hub

import (
	"context"
	"fmt"
	"strings"
)

// CombinedClusterClient provides operations that combine ManagedCluster and ClusterDeployment data
type CombinedClusterClient interface {
	// ListCombined fetches all ManagedClusters and enriches them with ClusterDeployment data
	ListCombined(ctx context.Context) ([]CombinedClusterInfo, error)
}

type combinedClusterClient struct {
	managedClusterClient     ManagedClusterClient
	clusterDeploymentClient  ClusterDeploymentClient
}

// NewCombinedClusterClient creates a new CombinedClusterClient
func NewCombinedClusterClient(
	mcClient ManagedClusterClient,
	cdClient ClusterDeploymentClient,
) CombinedClusterClient {
	return &combinedClusterClient{
		managedClusterClient:    mcClient,
		clusterDeploymentClient: cdClient,
	}
}

// ListCombined fetches all ManagedClusters and enriches them with ClusterDeployment data
// If a ClusterDeployment is not found for a ManagedCluster, it still includes the ManagedCluster
// data with default/N/A values for ClusterDeployment fields
func (c *combinedClusterClient) ListCombined(ctx context.Context) ([]CombinedClusterInfo, error) {
	// First, list all ManagedClusters
	managedClusters, err := c.managedClusterClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed clusters: %w", err)
	}

	// For each ManagedCluster, try to fetch the corresponding ClusterDeployment
	combined := make([]CombinedClusterInfo, 0, len(managedClusters))
	for _, mc := range managedClusters {
		info := CombinedClusterInfo{
			Name:      mc.Name,
			Status:    mc.Status,
			Available: mc.Available,
			Message:   mc.Message,
		}

		// Try to get ClusterDeployment data
		// ClusterDeployment is in namespace=cluster-name with name=cluster-name
		cd, err := c.clusterDeploymentClient.Get(ctx, mc.Name)
		if err != nil {
			// If ClusterDeployment not found (e.g., non-Hive cluster), use N/A values
			if isNotFoundError(err) {
				info.PowerState = "N/A"
				info.Platform = "N/A"
				info.Region = "N/A"
				info.Version = "N/A"
				info.APIUrl = ""
				info.ConsoleURL = ""
				info.KubeconfigSecret = ""
			} else {
				// For other errors, log but continue
				// In a real implementation, we might want to log this
				info.PowerState = "Unknown"
				info.Platform = "Unknown"
				info.Region = "Unknown"
				info.Version = "Unknown"
			}
		} else {
			// Merge ClusterDeployment data
			info.PowerState = cd.PowerState
			info.Platform = cd.Platform
			info.Region = cd.Region
			info.Version = cd.Version
			info.APIUrl = cd.APIUrl
			info.ConsoleURL = cd.ConsoleURL

			// Format kubeconfig secret as namespace/name
			if cd.KubeconfigSecretName != "" {
				info.KubeconfigSecret = fmt.Sprintf("%s/%s", cd.KubeconfigSecretNS, cd.KubeconfigSecretName)
			}
		}

		combined = append(combined, info)
	}

	return combined, nil
}

// isNotFoundError checks if an error is a "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}
