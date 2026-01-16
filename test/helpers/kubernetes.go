package helpers

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/yaml"
)

// CreateTestManagedCluster creates a test ManagedCluster with the specified name and availability status
func CreateTestManagedCluster(name string, available string) *clusterv1.ManagedCluster {
	cluster := &clusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
			},
		},
		Spec: clusterv1.ManagedClusterSpec{
			HubAcceptsClient:     true,
			LeaseDurationSeconds: 60,
		},
		Status: clusterv1.ManagedClusterStatus{
			Conditions: []metav1.Condition{},
		},
	}

	// Add availability condition based on the parameter
	var conditionStatus metav1.ConditionStatus
	var message string
	var reason string

	switch available {
	case "True":
		conditionStatus = metav1.ConditionTrue
		message = "Managed cluster is available"
		reason = "ManagedClusterAvailable"
	case "False":
		conditionStatus = metav1.ConditionFalse
		message = "Managed cluster is not available"
		reason = "ManagedClusterNotAvailable"
	case "Unknown":
		conditionStatus = metav1.ConditionUnknown
		message = "Managed cluster status is unknown"
		reason = "ManagedClusterUnknown"
	default:
		// Default to Unknown if invalid value provided
		conditionStatus = metav1.ConditionUnknown
		message = "Managed cluster status is unknown"
		reason = "ManagedClusterUnknown"
	}

	cluster.Status.Conditions = append(cluster.Status.Conditions, metav1.Condition{
		Type:               clusterv1.ManagedClusterConditionAvailable,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})

	return cluster
}

// LoadManagedClusterFromFile loads a ManagedCluster from a YAML file
func LoadManagedClusterFromFile(path string) (*clusterv1.ManagedCluster, error) {
	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Create a scheme with the ManagedCluster type registered
	scheme := runtime.NewScheme()
	if err := clusterv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add clusterv1 to scheme: %w", err)
	}

	// Create a decoder
	codecs := serializer.NewCodecFactory(scheme)
	decoder := codecs.UniversalDeserializer()

	// Decode the YAML
	obj, _, err := decoder.Decode(data, nil, nil)
	if err != nil {
		// Try using sigs.k8s.io/yaml as fallback
		cluster := &clusterv1.ManagedCluster{}
		if yamlErr := yaml.Unmarshal(data, cluster); yamlErr != nil {
			return nil, fmt.Errorf("failed to decode YAML (decoder: %v, yaml: %v)", err, yamlErr)
		}
		return cluster, nil
	}

	// Type assert to ManagedCluster
	cluster, ok := obj.(*clusterv1.ManagedCluster)
	if !ok {
		return nil, fmt.Errorf("decoded object is not a ManagedCluster, got %T", obj)
	}

	return cluster, nil
}

// LoadClusterDeploymentFromFile loads a ClusterDeployment as an unstructured object from a YAML file
// We use unstructured.Unstructured to avoid importing the full Hive API
func LoadClusterDeploymentFromFile(path string) (*unstructured.Unstructured, error) {
	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Parse YAML into unstructured format
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(data, obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return obj, nil
}
