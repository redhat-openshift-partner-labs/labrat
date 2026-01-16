package spoke_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/spoke"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("KubeconfigExtractor", func() {
	var (
		extractor       spoke.KubeconfigExtractor
		fakeK8s         *k8sFake.Clientset
		fakeDynamic     *fake.FakeDynamicClient
		ctx             context.Context
		clusterName     string
		validKubeconfig string
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterName = "test-cluster"
		validKubeconfig = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://api.test-cluster.example.com:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: admin
  name: admin
current-context: admin
users:
- name: admin
  user:
    token: test-token
`
	})

	Describe("Extract", func() {
		Context("with valid ClusterDeployment and Secret", func() {
			BeforeEach(func() {
				// Create a fake ClusterDeployment
				cd := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "hive.openshift.io/v1",
						"kind":       "ClusterDeployment",
						"metadata": map[string]interface{}{
							"name":      clusterName,
							"namespace": clusterName,
						},
						"spec": map[string]interface{}{
							"clusterMetadata": map[string]interface{}{
								"adminKubeconfigSecretRef": map[string]interface{}{
									"name": clusterName + "-admin-kubeconfig",
								},
							},
						},
					},
				}

				// Create a fake Secret with kubeconfig
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-admin-kubeconfig",
						Namespace: clusterName,
					},
					Data: map[string][]byte{
						"kubeconfig": []byte(validKubeconfig),
					},
				}

				// Setup fake clients
				scheme := runtime.NewScheme()
				fakeDynamic = fake.NewSimpleDynamicClient(scheme, cd)
				fakeK8s = k8sFake.NewSimpleClientset(secret)

				extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
			})

			It("should extract kubeconfig successfully", func() {
				kubeconfig, err := extractor.Extract(ctx, clusterName)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(kubeconfig)).To(Equal(validKubeconfig))
			})
		})

		Context("with base64-encoded kubeconfig in secret", func() {
			BeforeEach(func() {
				// Create a fake ClusterDeployment
				cd := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "hive.openshift.io/v1",
						"kind":       "ClusterDeployment",
						"metadata": map[string]interface{}{
							"name":      clusterName,
							"namespace": clusterName,
						},
						"spec": map[string]interface{}{
							"clusterMetadata": map[string]interface{}{
								"adminKubeconfigSecretRef": map[string]interface{}{
									"name": clusterName + "-admin-kubeconfig",
								},
							},
						},
					},
				}

				// Create a fake Secret with base64-encoded kubeconfig
				// Note: Kubernetes secrets are already base64-encoded when stored,
				// but we test both raw and double-encoded for robustness
				encodedKubeconfig := base64.StdEncoding.EncodeToString([]byte(validKubeconfig))
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-admin-kubeconfig",
						Namespace: clusterName,
					},
					Data: map[string][]byte{
						"kubeconfig": []byte(encodedKubeconfig),
					},
				}

				// Setup fake clients
				scheme := runtime.NewScheme()
				fakeDynamic = fake.NewSimpleDynamicClient(scheme, cd)
				fakeK8s = k8sFake.NewSimpleClientset(secret)

				extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
			})

			It("should handle base64-encoded kubeconfig", func() {
				kubeconfig, err := extractor.Extract(ctx, clusterName)
				Expect(err).NotTo(HaveOccurred())
				// Should return the decoded version
				Expect(string(kubeconfig)).To(Equal(validKubeconfig))
			})
		})

		Context("when ClusterDeployment is not found", func() {
			BeforeEach(func() {
				// Setup fake clients with no resources
				scheme := runtime.NewScheme()
				fakeDynamic = fake.NewSimpleDynamicClient(scheme)
				fakeK8s = k8sFake.NewSimpleClientset()

				extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
			})

			It("should return error", func() {
				_, err := extractor.Extract(ctx, clusterName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when Secret is not found", func() {
			BeforeEach(func() {
				// Create a fake ClusterDeployment
				cd := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "hive.openshift.io/v1",
						"kind":       "ClusterDeployment",
						"metadata": map[string]interface{}{
							"name":      clusterName,
							"namespace": clusterName,
						},
						"spec": map[string]interface{}{
							"clusterMetadata": map[string]interface{}{
								"adminKubeconfigSecretRef": map[string]interface{}{
									"name": clusterName + "-admin-kubeconfig",
								},
							},
						},
					},
				}

				// Setup fake clients (no secret)
				scheme := runtime.NewScheme()
				fakeDynamic = fake.NewSimpleDynamicClient(scheme, cd)
				fakeK8s = k8sFake.NewSimpleClientset()

				extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
			})

			It("should return error", func() {
				_, err := extractor.Extract(ctx, clusterName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when Secret is missing kubeconfig key", func() {
			BeforeEach(func() {
				// Create a fake ClusterDeployment
				cd := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "hive.openshift.io/v1",
						"kind":       "ClusterDeployment",
						"metadata": map[string]interface{}{
							"name":      clusterName,
							"namespace": clusterName,
						},
						"spec": map[string]interface{}{
							"clusterMetadata": map[string]interface{}{
								"adminKubeconfigSecretRef": map[string]interface{}{
									"name": clusterName + "-admin-kubeconfig",
								},
							},
						},
					},
				}

				// Create a fake Secret without kubeconfig key
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-admin-kubeconfig",
						Namespace: clusterName,
					},
					Data: map[string][]byte{
						"some-other-key": []byte("data"),
					},
				}

				// Setup fake clients
				scheme := runtime.NewScheme()
				fakeDynamic = fake.NewSimpleDynamicClient(scheme, cd)
				fakeK8s = k8sFake.NewSimpleClientset(secret)

				extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
			})

			It("should return error", func() {
				_, err := extractor.Extract(ctx, clusterName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubeconfig"))
			})
		})
	})

	Describe("ExtractToFile", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "labrat-test-*")
			Expect(err).NotTo(HaveOccurred())

			// Create a fake ClusterDeployment
			cd := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "hive.openshift.io/v1",
					"kind":       "ClusterDeployment",
					"metadata": map[string]interface{}{
						"name":      clusterName,
						"namespace": clusterName,
					},
					"spec": map[string]interface{}{
						"clusterMetadata": map[string]interface{}{
							"adminKubeconfigSecretRef": map[string]interface{}{
								"name": clusterName + "-admin-kubeconfig",
							},
						},
					},
				},
			}

			// Create a fake Secret with kubeconfig
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName + "-admin-kubeconfig",
					Namespace: clusterName,
				},
				Data: map[string][]byte{
					"kubeconfig": []byte(validKubeconfig),
				},
			}

			// Setup fake clients
			scheme := runtime.NewScheme()
			fakeDynamic = fake.NewSimpleDynamicClient(scheme, cd)
			fakeK8s = k8sFake.NewSimpleClientset(secret)

			extractor = spoke.NewKubeconfigExtractor(fakeDynamic, fakeK8s.CoreV1())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("should write kubeconfig to file successfully", func() {
			outputPath := filepath.Join(tmpDir, "kubeconfig")
			err := extractor.ExtractToFile(ctx, clusterName, outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify file contents
			content, err := os.ReadFile(outputPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(validKubeconfig))
		})

		It("should set restrictive file permissions (0600)", func() {
			outputPath := filepath.Join(tmpDir, "kubeconfig")
			err := extractor.ExtractToFile(ctx, clusterName, outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify file permissions
			info, err := os.Stat(outputPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
		})

		It("should create parent directories if needed", func() {
			outputPath := filepath.Join(tmpDir, "subdir", "nested", "kubeconfig")
			err := extractor.ExtractToFile(ctx, clusterName, outputPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify file exists
			_, err = os.Stat(outputPath)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
