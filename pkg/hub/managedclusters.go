package hub

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

const (
	// UnreachableTaintKey is the taint key for unreachable clusters
	UnreachableTaintKey = "cluster.open-cluster-management.io/unreachable"
)

// ManagedClusterClient provides methods to interact with ManagedCluster resources
type ManagedClusterClient interface {
	// List retrieves all managed clusters from the hub
	List(ctx context.Context) ([]ManagedClusterInfo, error)
	// Filter filters clusters based on the provided criteria
	Filter(clusters []ManagedClusterInfo, filter ManagedClusterFilter) []ManagedClusterInfo
}

type managedClusterClient struct {
	dynamicClient dynamic.Interface
}

// NewManagedClusterClient creates a new ManagedClusterClient
func NewManagedClusterClient(dynamicClient dynamic.Interface) ManagedClusterClient {
	return &managedClusterClient{
		dynamicClient: dynamicClient,
	}
}

// List retrieves all managed clusters from the hub and returns their information
func (m *managedClusterClient) List(ctx context.Context) ([]ManagedClusterInfo, error) {
	// Define the GVR for ManagedCluster
	gvr := schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}

	// List all ManagedCluster resources
	unstructuredList, err := m.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list managed clusters: %w", err)
	}

	var clusters []ManagedClusterInfo

	for _, item := range unstructuredList.Items {
		// Convert unstructured to ManagedCluster
		var cluster clusterv1.ManagedCluster
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &cluster)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured to ManagedCluster: %w", err)
		}

		// Extract cluster information
		info := ManagedClusterInfo{
			Name:   cluster.Name,
			Status: deriveStatus(&cluster),
		}

		// Get available condition
		info.Available, info.Message = getAvailableCondition(&cluster)

		clusters = append(clusters, info)
	}

	return clusters, nil
}

// Filter filters the list of clusters based on the provided filter criteria
func (m *managedClusterClient) Filter(clusters []ManagedClusterInfo, filter ManagedClusterFilter) []ManagedClusterInfo {
	// If no status filter is specified, return all clusters
	if filter.Status == "" {
		return clusters
	}

	var filtered []ManagedClusterInfo
	for _, cluster := range clusters {
		if cluster.Status == filter.Status {
			filtered = append(filtered, cluster)
		}
	}

	return filtered
}

// deriveStatus determines the overall status of a managed cluster
// Priority:
// 1. Check for unreachable taint → NotReady
// 2. Check ManagedClusterConditionAvailable:
//   - True → Ready
//   - False → NotReady
//   - Unknown → Unknown
//
// 3. Default → Unknown
func deriveStatus(cluster *clusterv1.ManagedCluster) ClusterStatus {
	// Check for unreachable taint first
	for _, taint := range cluster.Spec.Taints {
		if taint.Key == UnreachableTaintKey {
			return StatusNotReady
		}
	}

	// Check ManagedClusterConditionAvailable
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1.ManagedClusterConditionAvailable {
			switch condition.Status {
			case metav1.ConditionTrue:
				return StatusReady
			case metav1.ConditionFalse:
				return StatusNotReady
			case metav1.ConditionUnknown:
				return StatusUnknown
			}
		}
	}

	// Default to Unknown if no conditions are present
	return StatusUnknown
}

// getAvailableCondition extracts the Available condition status and message
func getAvailableCondition(cluster *clusterv1.ManagedCluster) (string, string) {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == clusterv1.ManagedClusterConditionAvailable {
			return string(condition.Status), condition.Message
		}
	}
	return "Unknown", ""
}
