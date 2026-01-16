// Package spoke provides functionality for managing and interacting with spoke clusters
package spoke

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// KubeconfigExtractor provides methods to extract admin kubeconfig from spoke clusters
type KubeconfigExtractor interface {
	// Extract retrieves the admin kubeconfig for a spoke cluster and returns it as bytes
	Extract(ctx context.Context, clusterName string) ([]byte, error)
	// ExtractToFile retrieves the admin kubeconfig and writes it to a file with secure permissions
	ExtractToFile(ctx context.Context, clusterName, outputPath string) error
}

type kubeconfigExtractor struct {
	dynamicClient dynamic.Interface
	coreClient    corev1.CoreV1Interface
}

// NewKubeconfigExtractor creates a new KubeconfigExtractor
func NewKubeconfigExtractor(
	dynamicClient dynamic.Interface,
	coreClient corev1.CoreV1Interface,
) KubeconfigExtractor {
	return &kubeconfigExtractor{
		dynamicClient: dynamicClient,
		coreClient:    coreClient,
	}
}

// Extract retrieves the admin kubeconfig for a spoke cluster
// Algorithm:
// 1. Get ClusterDeployment from namespace=clusterName, name=clusterName
// 2. Extract spec.clusterMetadata.adminKubeconfigSecretRef.name
// 3. Get Secret from namespace=clusterName, name=secretName
// 4. Extract data["kubeconfig"]
// 5. Decode if base64 encoded (beyond Kubernetes' native encoding)
// 6. Validate and return
func (k *kubeconfigExtractor) Extract(ctx context.Context, clusterName string) ([]byte, error) {
	// Step 1: Get ClusterDeployment
	gvr := schema.GroupVersionResource{
		Group:    "hive.openshift.io",
		Version:  "v1",
		Resource: "clusterdeployments",
	}

	cd, err := k.dynamicClient.Resource(gvr).Namespace(clusterName).Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterDeployment %s: %w (cluster not found or not managed by Hive)", clusterName, err)
	}

	// Step 2: Extract secret reference
	spec, ok := cd.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ClusterDeployment spec not found")
	}

	clusterMetadata, ok := spec["clusterMetadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ClusterDeployment clusterMetadata not found")
	}

	kubeconfigRef, ok := clusterMetadata["adminKubeconfigSecretRef"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("adminKubeconfigSecretRef not found in ClusterDeployment")
	}

	secretName, ok := kubeconfigRef["name"].(string)
	if !ok {
		return nil, fmt.Errorf("secret name not found in adminKubeconfigSecretRef")
	}

	// Step 3: Get Secret
	secret, err := k.coreClient.Secrets(clusterName).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin kubeconfig secret %s/%s: %w", clusterName, secretName, err)
	}

	// Step 4: Extract kubeconfig data
	kubeconfigData, ok := secret.Data["kubeconfig"]
	if !ok {
		return nil, fmt.Errorf("kubeconfig key not found in secret %s/%s", clusterName, secretName)
	}

	if len(kubeconfigData) == 0 {
		return nil, fmt.Errorf("kubeconfig data is empty in secret %s/%s", clusterName, secretName)
	}

	// Step 5: Check if data is double-encoded (base64 on top of Kubernetes' native encoding)
	// This is a common pattern in some environments
	kubeconfig := kubeconfigData
	if isBase64Encoded(kubeconfigData) {
		decoded, err := base64.StdEncoding.DecodeString(string(kubeconfigData))
		if err == nil {
			// Successfully decoded, use decoded version
			kubeconfig = decoded
		}
		// If decoding fails, assume it's already raw YAML
	}

	// Step 6: Basic validation - check for YAML structure
	kubeconfigStr := string(kubeconfig)
	if !strings.Contains(kubeconfigStr, "apiVersion:") || !strings.Contains(kubeconfigStr, "kind:") {
		return nil, fmt.Errorf("kubeconfig validation failed: missing required YAML fields")
	}

	return kubeconfig, nil
}

// ExtractToFile extracts the kubeconfig and writes it to a file with secure permissions
func (k *kubeconfigExtractor) ExtractToFile(ctx context.Context, clusterName, outputPath string) error {
	// Extract kubeconfig
	kubeconfig, err := k.Extract(ctx, clusterName)
	if err != nil {
		return err
	}

	// Create parent directories if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file with restrictive permissions (0600)
	if err := os.WriteFile(outputPath, kubeconfig, 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig to %s: %w", outputPath, err)
	}

	return nil
}

// isBase64Encoded checks if data appears to be base64 encoded
// Heuristic: if it's valid base64 and doesn't look like YAML, it's probably encoded
func isBase64Encoded(data []byte) bool {
	s := string(data)

	// If it looks like YAML already, it's not base64-encoded
	if strings.HasPrefix(strings.TrimSpace(s), "apiVersion:") {
		return false
	}

	// Check if it's valid base64
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
