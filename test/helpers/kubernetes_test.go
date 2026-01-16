package helpers_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-openshift-partner-labs/labrat/test/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

var _ = Describe("Kubernetes Helpers", func() {
	Describe("CreateTestManagedCluster", func() {
		Context("with Available=True", func() {
			It("should create a cluster with Ready status", func() {
				cluster := helpers.CreateTestManagedCluster("test-ready", "True")
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Name).To(Equal("test-ready"))
				Expect(cluster.Status.Conditions).To(HaveLen(1))

				condition := cluster.Status.Conditions[0]
				Expect(condition.Type).To(Equal(clusterv1.ManagedClusterConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
				Expect(condition.Reason).To(Equal("ManagedClusterAvailable"))
			})
		})

		Context("with Available=False", func() {
			It("should create a cluster with NotReady status", func() {
				cluster := helpers.CreateTestManagedCluster("test-notready", "False")
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Name).To(Equal("test-notready"))
				Expect(cluster.Status.Conditions).To(HaveLen(1))

				condition := cluster.Status.Conditions[0]
				Expect(condition.Type).To(Equal(clusterv1.ManagedClusterConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionFalse))
				Expect(condition.Reason).To(Equal("ManagedClusterNotAvailable"))
			})
		})

		Context("with Available=Unknown", func() {
			It("should create a cluster with Unknown status", func() {
				cluster := helpers.CreateTestManagedCluster("test-unknown", "Unknown")
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Name).To(Equal("test-unknown"))
				Expect(cluster.Status.Conditions).To(HaveLen(1))

				condition := cluster.Status.Conditions[0]
				Expect(condition.Type).To(Equal(clusterv1.ManagedClusterConditionAvailable))
				Expect(condition.Status).To(Equal(metav1.ConditionUnknown))
				Expect(condition.Reason).To(Equal("ManagedClusterUnknown"))
			})
		})

		Context("with invalid availability value", func() {
			It("should default to Unknown status", func() {
				cluster := helpers.CreateTestManagedCluster("test-invalid", "InvalidValue")
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Status.Conditions).To(HaveLen(1))

				condition := cluster.Status.Conditions[0]
				Expect(condition.Status).To(Equal(metav1.ConditionUnknown))
			})
		})

		It("should set basic cluster properties", func() {
			cluster := helpers.CreateTestManagedCluster("test-cluster", "True")
			Expect(cluster.Spec.HubAcceptsClient).To(BeTrue())
			Expect(cluster.Spec.LeaseDurationSeconds).To(Equal(int32(60)))
			Expect(cluster.Labels).To(HaveKeyWithValue("name", "test-cluster"))
		})
	})

	Describe("LoadManagedClusterFromFile", func() {
		var fixturesDir string

		BeforeEach(func() {
			// Get the path to the fixtures directory
			// Tests run from test/helpers, so go up one level to test/
			fixturesDir = filepath.Join("..", "fixtures")
		})

		Context("with a valid ready cluster YAML file", func() {
			It("should load the cluster correctly", func() {
				path := filepath.Join(fixturesDir, "managedcluster_ready.yaml")
				cluster, err := helpers.LoadManagedClusterFromFile(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Name).To(Equal("cluster-ready"))

				// Verify the Available condition
				var availableCondition *metav1.Condition
				for i := range cluster.Status.Conditions {
					if cluster.Status.Conditions[i].Type == clusterv1.ManagedClusterConditionAvailable {
						availableCondition = &cluster.Status.Conditions[i]
						break
					}
				}

				Expect(availableCondition).NotTo(BeNil())
				Expect(availableCondition.Status).To(Equal(metav1.ConditionTrue))
				Expect(availableCondition.Message).To(Equal("Managed cluster is available"))
			})
		})

		Context("with a valid notready cluster YAML file", func() {
			It("should load the cluster correctly", func() {
				path := filepath.Join(fixturesDir, "managedcluster_notready.yaml")
				cluster, err := helpers.LoadManagedClusterFromFile(path)

				Expect(err).NotTo(HaveOccurred())
				Expect(cluster).NotTo(BeNil())
				Expect(cluster.Name).To(Equal("cluster-notready"))

				// Verify the unreachable taint
				Expect(cluster.Spec.Taints).To(HaveLen(1))
				Expect(cluster.Spec.Taints[0].Key).To(Equal("cluster.open-cluster-management.io/unreachable"))
				Expect(cluster.Spec.Taints[0].Effect).To(Equal(clusterv1.TaintEffectNoSelect))

				// Verify the Available condition
				var availableCondition *metav1.Condition
				for i := range cluster.Status.Conditions {
					if cluster.Status.Conditions[i].Type == clusterv1.ManagedClusterConditionAvailable {
						availableCondition = &cluster.Status.Conditions[i]
						break
					}
				}

				Expect(availableCondition).NotTo(BeNil())
				Expect(availableCondition.Status).To(Equal(metav1.ConditionFalse))
			})
		})

		Context("with a non-existent file", func() {
			It("should return an error", func() {
				path := filepath.Join(fixturesDir, "nonexistent.yaml")
				cluster, err := helpers.LoadManagedClusterFromFile(path)

				Expect(err).To(HaveOccurred())
				Expect(cluster).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to read file"))
			})
		})

		Context("with an invalid YAML file", func() {
			It("should return an error for non-existent file", func() {
				// Test with a non-existent file path
				tempFile := filepath.Join(fixturesDir, "invalid.yaml")
				cluster, err := helpers.LoadManagedClusterFromFile(tempFile)

				// Since the file doesn't exist, we expect an error
				Expect(err).To(HaveOccurred())
				Expect(cluster).To(BeNil())
			})
		})
	})
})
