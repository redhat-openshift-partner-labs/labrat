//go:build test

package hub_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
	"github.com/redhat-openshift-partner-labs/labrat/test/helpers"
)

var _ = Describe("CombinedClusterClient", func() {
	var (
		client       hub.CombinedClusterClient
		mockMCClient *mockManagedClusterClientForCombined
		mockCDClient *mockClusterDeploymentClientForCombined
	)

	BeforeEach(func() {
		mockMCClient = newMockManagedClusterClientForCombined()
		mockCDClient = newMockClusterDeploymentClientForCombined()
		client = hub.NewCombinedClusterClient(mockMCClient, mockCDClient)
	})

	Describe("ListCombined", func() {
		Context("when both ManagedCluster and ClusterDeployment exist", func() {
			It("should merge data from both resources", func() {
				// Setup ManagedCluster
				mc, err := helpers.LoadManagedClusterFromFile("../../test/fixtures/managedcluster_ready.yaml")
				Expect(err).NotTo(HaveOccurred())

				mockMCClient.managedClusters = []hub.ManagedClusterInfo{
					{
						Name:      mc.Name,
						Status:    hub.StatusReady,
						Available: "True",
						Message:   "Cluster is available",
					},
				}

				// Setup ClusterDeployment
				cd, err := helpers.LoadClusterDeploymentFromFile("../../test/fixtures/clusterdeployment_running.yaml")
				Expect(err).NotTo(HaveOccurred())

				mockCDClient.clusterDeployments = map[string]*hub.ClusterDeploymentInfo{
					"cluster-ready": {
						Name:                 cd.GetName(),
						Namespace:            cd.GetNamespace(),
						PowerState:           "Running",
						Installed:            true,
						APIUrl:               "https://api.test-cluster-running.example.com:6443",
						ConsoleURL:           "https://console.test-cluster-running.example.com",
						KubeconfigSecretName: "test-cluster-running-admin-kubeconfig",
						KubeconfigSecretNS:   "test-cluster-running",
						Platform:             "aws",
						Region:               "us-east-1",
						Version:              "4.20.6",
					},
				}

				combined, err := client.ListCombined(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(combined).To(HaveLen(1))

				cluster := combined[0]
				Expect(cluster.Name).To(Equal("cluster-ready"))
				Expect(cluster.Status).To(Equal(hub.StatusReady))
				Expect(cluster.Available).To(Equal("True"))
				Expect(cluster.PowerState).To(Equal("Running"))
				Expect(cluster.Platform).To(Equal("aws"))
				Expect(cluster.Region).To(Equal("us-east-1"))
				Expect(cluster.Version).To(Equal("4.20.6"))
				Expect(cluster.APIUrl).To(Equal("https://api.test-cluster-running.example.com:6443"))
				Expect(cluster.ConsoleURL).To(Equal("https://console.test-cluster-running.example.com"))
				Expect(cluster.KubeconfigSecret).To(Equal("test-cluster-running/test-cluster-running-admin-kubeconfig"))
			})
		})

		Context("when ClusterDeployment is not found", func() {
			It("should still return ManagedCluster data with empty ClusterDeployment fields", func() {
				mockMCClient.managedClusters = []hub.ManagedClusterInfo{
					{
						Name:      "test-cluster",
						Status:    hub.StatusReady,
						Available: "True",
						Message:   "Cluster is available",
					},
				}

				// No ClusterDeployment data
				mockCDClient.clusterDeployments = map[string]*hub.ClusterDeploymentInfo{}

				combined, err := client.ListCombined(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(combined).To(HaveLen(1))

				cluster := combined[0]
				Expect(cluster.Name).To(Equal("test-cluster"))
				Expect(cluster.Status).To(Equal(hub.StatusReady))
				Expect(cluster.PowerState).To(Equal("N/A"))
				Expect(cluster.Platform).To(Equal("N/A"))
				Expect(cluster.Version).To(Equal("N/A"))
			})
		})

		Context("when no managed clusters exist", func() {
			It("should return empty list", func() {
				mockMCClient.managedClusters = []hub.ManagedClusterInfo{}
				mockCDClient.clusterDeployments = map[string]*hub.ClusterDeploymentInfo{}

				combined, err := client.ListCombined(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(combined).To(HaveLen(0))
			})
		})
	})
})

// Mock implementations for combined client testing
type mockManagedClusterClientForCombined struct {
	managedClusters []hub.ManagedClusterInfo
}

func newMockManagedClusterClientForCombined() *mockManagedClusterClientForCombined {
	return &mockManagedClusterClientForCombined{
		managedClusters: []hub.ManagedClusterInfo{},
	}
}

func (m *mockManagedClusterClientForCombined) List(ctx context.Context) ([]hub.ManagedClusterInfo, error) {
	return m.managedClusters, nil
}

func (m *mockManagedClusterClientForCombined) Filter(clusters []hub.ManagedClusterInfo, filter hub.ManagedClusterFilter) []hub.ManagedClusterInfo {
	return clusters
}

type mockClusterDeploymentClientForCombined struct {
	clusterDeployments map[string]*hub.ClusterDeploymentInfo
}

func newMockClusterDeploymentClientForCombined() *mockClusterDeploymentClientForCombined {
	return &mockClusterDeploymentClientForCombined{
		clusterDeployments: make(map[string]*hub.ClusterDeploymentInfo),
	}
}

func (m *mockClusterDeploymentClientForCombined) Get(ctx context.Context, name string) (*hub.ClusterDeploymentInfo, error) {
	if cd, ok := m.clusterDeployments[name]; ok {
		return cd, nil
	}
	// Return NotFound error
	return nil, &clusterDeploymentNotFoundError{name: name}
}

type clusterDeploymentNotFoundError struct {
	name string
}

func (e *clusterDeploymentNotFoundError) Error() string {
	return "clusterdeployment.hive.openshift.io \"" + e.name + "\" not found"
}
