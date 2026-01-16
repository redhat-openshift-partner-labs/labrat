package hub_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// Mock dynamic client implementation
type mockDynamicClient struct {
	clusters []clusterv1.ManagedCluster
}

type mockResourceInterface struct {
	clusters []clusterv1.ManagedCluster
}

func (m *mockDynamicClient) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &mockResourceInterface{clusters: m.clusters}
}

func (m *mockResourceInterface) Namespace(string) dynamic.ResourceInterface {
	return m
}

func (m *mockResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{}
	for _, cluster := range m.clusters {
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cluster)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, unstructured.Unstructured{Object: unstructuredObj})
	}
	return list, nil
}

func (m *mockResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (m *mockResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (m *mockResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m *mockResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (m *mockResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

var _ = Describe("ManagedClusterClient", func() {
	var (
		dynamicClient dynamic.Interface
		client        hub.ManagedClusterClient
		ctx           context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("List", func() {
		Context("with no clusters", func() {
			BeforeEach(func() {
				dynamicClient = &mockDynamicClient{clusters: []clusterv1.ManagedCluster{}}
				client = hub.NewManagedClusterClient(dynamicClient)
			})

			It("should return empty list", func() {
				clusters, err := client.List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(BeEmpty())
			})
		})

		Context("with multiple clusters", func() {
			BeforeEach(func() {
				readyCluster := clusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-ready",
					},
					Status: clusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:    clusterv1.ManagedClusterConditionAvailable,
								Status:  metav1.ConditionTrue,
								Message: "Cluster is available",
							},
						},
					},
				}

				notReadyCluster := clusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-notready",
					},
					Status: clusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:    clusterv1.ManagedClusterConditionAvailable,
								Status:  metav1.ConditionFalse,
								Message: "Cluster is not available",
							},
						},
					},
				}

				unknownCluster := clusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-unknown",
					},
					Status: clusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:    clusterv1.ManagedClusterConditionAvailable,
								Status:  metav1.ConditionUnknown,
								Message: "Cluster status unknown",
							},
						},
					},
				}

				dynamicClient = &mockDynamicClient{
					clusters: []clusterv1.ManagedCluster{readyCluster, notReadyCluster, unknownCluster},
				}
				client = hub.NewManagedClusterClient(dynamicClient)
			})

			It("should return all clusters with correct status", func() {
				clusters, err := client.List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(3))

				clusterMap := make(map[string]hub.ManagedClusterInfo)
				for _, c := range clusters {
					clusterMap[c.Name] = c
				}

				Expect(clusterMap["cluster-ready"].Status).To(Equal(hub.StatusReady))
				Expect(clusterMap["cluster-ready"].Available).To(Equal("True"))

				Expect(clusterMap["cluster-notready"].Status).To(Equal(hub.StatusNotReady))
				Expect(clusterMap["cluster-notready"].Available).To(Equal("False"))

				Expect(clusterMap["cluster-unknown"].Status).To(Equal(hub.StatusUnknown))
				Expect(clusterMap["cluster-unknown"].Available).To(Equal("Unknown"))
			})
		})

		Context("with cluster having unreachable taint", func() {
			BeforeEach(func() {
				unreachableCluster := clusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-unreachable",
					},
					Spec: clusterv1.ManagedClusterSpec{
						Taints: []clusterv1.Taint{
							{
								Key:    "cluster.open-cluster-management.io/unreachable",
								Effect: clusterv1.TaintEffectNoSelect,
							},
						},
					},
					Status: clusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:   clusterv1.ManagedClusterConditionAvailable,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}

				dynamicClient = &mockDynamicClient{clusters: []clusterv1.ManagedCluster{unreachableCluster}}
				client = hub.NewManagedClusterClient(dynamicClient)
			})

			It("should mark cluster as NotReady due to unreachable taint", func() {
				clusters, err := client.List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Status).To(Equal(hub.StatusNotReady))
			})
		})

		Context("with cluster having no conditions", func() {
			BeforeEach(func() {
				noConditionsCluster := clusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-no-conditions",
					},
					Status: clusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{},
					},
				}

				dynamicClient = &mockDynamicClient{clusters: []clusterv1.ManagedCluster{noConditionsCluster}}
				client = hub.NewManagedClusterClient(dynamicClient)
			})

			It("should mark cluster as Unknown", func() {
				clusters, err := client.List(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Status).To(Equal(hub.StatusUnknown))
				Expect(clusters[0].Available).To(Equal("Unknown"))
			})
		})
	})

	Describe("Filter", func() {
		var clusters []hub.ManagedClusterInfo

		BeforeEach(func() {
			clusters = []hub.ManagedClusterInfo{
				{Name: "cluster-1", Status: hub.StatusReady, Available: "True"},
				{Name: "cluster-2", Status: hub.StatusNotReady, Available: "False"},
				{Name: "cluster-3", Status: hub.StatusReady, Available: "True"},
				{Name: "cluster-4", Status: hub.StatusUnknown, Available: "Unknown"},
				{Name: "cluster-5", Status: hub.StatusNotReady, Available: "False"},
			}

			dynamicClient = &mockDynamicClient{clusters: []clusterv1.ManagedCluster{}}
			client = hub.NewManagedClusterClient(dynamicClient)
		})

		Context("filtering by Ready status", func() {
			It("should return only Ready clusters", func() {
				filter := hub.ManagedClusterFilter{Status: hub.StatusReady}
				filtered := client.Filter(clusters, filter)
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].Name).To(Equal("cluster-1"))
				Expect(filtered[1].Name).To(Equal("cluster-3"))
			})
		})

		Context("filtering by NotReady status", func() {
			It("should return only NotReady clusters", func() {
				filter := hub.ManagedClusterFilter{Status: hub.StatusNotReady}
				filtered := client.Filter(clusters, filter)
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].Name).To(Equal("cluster-2"))
				Expect(filtered[1].Name).To(Equal("cluster-5"))
			})
		})

		Context("filtering by Unknown status", func() {
			It("should return only Unknown clusters", func() {
				filter := hub.ManagedClusterFilter{Status: hub.StatusUnknown}
				filtered := client.Filter(clusters, filter)
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].Name).To(Equal("cluster-4"))
			})
		})

		Context("with empty filter", func() {
			It("should return all clusters", func() {
				filter := hub.ManagedClusterFilter{}
				filtered := client.Filter(clusters, filter)
				Expect(filtered).To(HaveLen(5))
			})
		})

		Context("with empty cluster list", func() {
			It("should return empty list", func() {
				filter := hub.ManagedClusterFilter{Status: hub.StatusReady}
				filtered := client.Filter([]hub.ManagedClusterInfo{}, filter)
				Expect(filtered).To(BeEmpty())
			})
		})
	})
})
