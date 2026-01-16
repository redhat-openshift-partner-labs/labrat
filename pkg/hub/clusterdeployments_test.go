//go:build test

package hub_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
	"github.com/redhat-openshift-partner-labs/labrat/test/helpers"
)

var _ = Describe("ClusterDeploymentClient", func() {
	var (
		client hub.ClusterDeploymentClient
		mockDynamicClient *mockDynamicClientForCD
	)

	BeforeEach(func() {
		mockDynamicClient = newMockDynamicClientForCD()
		client = hub.NewClusterDeploymentClient(mockDynamicClient)
	})

	Describe("Get", func() {
		Context("when ClusterDeployment exists", func() {
			It("should return ClusterDeployment info for a running cluster", func() {
				cd, err := helpers.LoadClusterDeploymentFromFile("../../test/fixtures/clusterdeployment_running.yaml")
				Expect(err).NotTo(HaveOccurred())

				mockDynamicClient.clusterDeployments["test-cluster-running"] = cd

				info, err := client.Get(context.Background(), "test-cluster-running")
				Expect(err).NotTo(HaveOccurred())
				Expect(info).NotTo(BeNil())
				Expect(info.Name).To(Equal("test-cluster-running"))
				Expect(info.Namespace).To(Equal("test-cluster-running"))
				Expect(info.PowerState).To(Equal("Running"))
				Expect(info.Installed).To(BeTrue())
				Expect(info.APIUrl).To(Equal("https://api.test-cluster-running.example.com:6443"))
				Expect(info.ConsoleURL).To(Equal("https://console.test-cluster-running.example.com"))
				Expect(info.KubeconfigSecretName).To(Equal("test-cluster-running-admin-kubeconfig"))
				Expect(info.KubeconfigSecretNS).To(Equal("test-cluster-running"))
				Expect(info.Platform).To(Equal("aws"))
				Expect(info.Region).To(Equal("us-east-1"))
				Expect(info.Version).To(Equal("4.20.6"))
			})

			It("should return ClusterDeployment info for a hibernating cluster", func() {
				cd, err := helpers.LoadClusterDeploymentFromFile("../../test/fixtures/clusterdeployment_hibernating.yaml")
				Expect(err).NotTo(HaveOccurred())

				mockDynamicClient.clusterDeployments["test-cluster-hibernating"] = cd

				info, err := client.Get(context.Background(), "test-cluster-hibernating")
				Expect(err).NotTo(HaveOccurred())
				Expect(info).NotTo(BeNil())
				Expect(info.Name).To(Equal("test-cluster-hibernating"))
				Expect(info.PowerState).To(Equal("Hibernating"))
				Expect(info.Version).To(Equal("4.19.8"))
			})
		})

		Context("when ClusterDeployment does not exist", func() {
			It("should return NotFound error", func() {
				info, err := client.Get(context.Background(), "nonexistent-cluster")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(info).To(BeNil())
			})
		})
	})
})

// Minimal mock for ClusterDeployment testing
type mockDynamicClientForCD struct {
	clusterDeployments map[string]*unstructured.Unstructured
}

func newMockDynamicClientForCD() *mockDynamicClientForCD {
	return &mockDynamicClientForCD{
		clusterDeployments: make(map[string]*unstructured.Unstructured),
	}
}

func (m *mockDynamicClientForCD) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &mockNamespaceableResourceForCD{
		client: m,
	}
}

type mockNamespaceableResourceForCD struct {
	client    *mockDynamicClientForCD
	namespace string
}

func (m *mockNamespaceableResourceForCD) Namespace(ns string) dynamic.ResourceInterface {
	return &mockResourceForCD{
		client:    m.client,
		namespace: ns,
	}
}

// Implement only the required methods for NamespaceableResourceInterface
func (m *mockNamespaceableResourceForCD) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *mockNamespaceableResourceForCD) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *mockNamespaceableResourceForCD) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockNamespaceableResourceForCD) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

type mockResourceForCD struct {
	client    *mockDynamicClientForCD
	namespace string
}

func (m *mockResourceForCD) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceForCD) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceForCD) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceForCD) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *mockResourceForCD) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *mockResourceForCD) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	if cd, ok := m.client.clusterDeployments[name]; ok {
		return cd, nil
	}
	return nil, fmt.Errorf("clusterdeployment.hive.openshift.io \"%s\" not found", name)
}

func (m *mockResourceForCD) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, nil
}

func (m *mockResourceForCD) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *mockResourceForCD) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceForCD) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceForCD) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}
