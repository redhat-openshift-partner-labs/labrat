// Package kube provides Kubernetes client initialization and management.
// It handles kubeconfig loading, context selection, and dynamic client creation
// for interacting with Kubernetes resources, including Custom Resource Definitions (CRDs).
package kube

import (
	"fmt"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client provides access to Kubernetes API via dynamic client
type Client struct {
	config  *rest.Config
	dynamic dynamic.Interface
}

// NewClient creates a new Kubernetes client from the specified kubeconfig file
// If context is empty, the current context from the kubeconfig will be used
func NewClient(kubeconfigPath string, context string) (*Client, error) {
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("kubeconfig path cannot be empty")
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfigPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfigPath)
		}
		return nil, fmt.Errorf("failed to access kubeconfig file: %w", err)
	}

	// Load kubeconfig
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeconfigPath,
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Build rest.Config
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		config:  config,
		dynamic: dynamicClient,
	}, nil
}

// GetDynamicClient returns the dynamic client interface for accessing Kubernetes resources
func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamic
}
