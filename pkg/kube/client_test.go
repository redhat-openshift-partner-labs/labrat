package kube_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/kube"
)

var _ = Describe("Client", func() {
	var (
		tempDir         string
		validKubeconfig string
		err             error
	)

	BeforeEach(func() {
		tempDir, err = os.MkdirTemp("", "kube-test-*")
		Expect(err).NotTo(HaveOccurred())

		validKubeconfig = filepath.Join(tempDir, "kubeconfig")
		kubeconfigContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-cluster:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
- context:
    cluster: test-cluster
    user: test-user
  name: another-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
		err = os.WriteFile(validKubeconfig, []byte(kubeconfigContent), 0600)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("NewClient", func() {
		Context("with valid kubeconfig", func() {
			It("should create a client successfully", func() {
				client, err := kube.NewClient(validKubeconfig, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("should return a client with dynamic interface", func() {
				client, err := kube.NewClient(validKubeconfig, "")
				Expect(err).NotTo(HaveOccurred())

				dynamicClient := client.GetDynamicClient()
				Expect(dynamicClient).NotTo(BeNil())
			})
		})

		Context("with specific context", func() {
			It("should use the specified context", func() {
				client, err := kube.NewClient(validKubeconfig, "another-context")
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})

		Context("with invalid kubeconfig path", func() {
			It("should return an error for non-existent file", func() {
				client, err := kube.NewClient("/nonexistent/kubeconfig", "")
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})
		})

		Context("with invalid kubeconfig content", func() {
			It("should return an error for malformed YAML", func() {
				invalidKubeconfig := filepath.Join(tempDir, "invalid-kubeconfig")
				err := os.WriteFile(invalidKubeconfig, []byte("invalid: yaml: content: ["), 0600)
				Expect(err).NotTo(HaveOccurred())

				client, err := kube.NewClient(invalidKubeconfig, "")
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})
		})

		Context("with non-existent context", func() {
			It("should return an error", func() {
				client, err := kube.NewClient(validKubeconfig, "non-existent-context")
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})
		})

		Context("with empty kubeconfig path", func() {
			It("should return an error", func() {
				client, err := kube.NewClient("", "")
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
			})
		})
	})

	Describe("GetDynamicClient", func() {
		It("should return a non-nil dynamic client", func() {
			client, err := kube.NewClient(validKubeconfig, "")
			Expect(err).NotTo(HaveOccurred())

			dynamicClient := client.GetDynamicClient()
			Expect(dynamicClient).NotTo(BeNil())
		})
	})
})
