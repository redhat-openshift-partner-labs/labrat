//go:build test

package config_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-openshift-partner-labs/labrat/internal/config"
)

var _ = Describe("Config", func() {
	var (
		tempDir    string
		configPath string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "labrat-test-")
		Expect(err).NotTo(HaveOccurred())
		configPath = filepath.Join(tempDir, "config.yaml")
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Loading Configuration", func() {
		Context("when a valid config file exists", func() {
			BeforeEach(func() {
				validConfig := `
hub:
  kubeconfig: /home/user/.kube/config
  context: hub-cluster
  namespace: open-cluster-management

defaults:
  spoke:
    provider: aws
    region: us-east-1

verbose: false
`
				err := os.WriteFile(configPath, []byte(validConfig), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should successfully load the configuration", func() {
				cfg, err := config.Load(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())
			})

			It("should parse hub configuration correctly", func() {
				cfg, err := config.Load(configPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Hub.Kubeconfig).To(Equal("/home/user/.kube/config"))
				Expect(cfg.Hub.Context).To(Equal("hub-cluster"))
				Expect(cfg.Hub.Namespace).To(Equal("open-cluster-management"))
			})

			It("should parse default spoke configuration", func() {
				cfg, err := config.Load(configPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Defaults.Spoke.Provider).To(Equal("aws"))
				Expect(cfg.Defaults.Spoke.Region).To(Equal("us-east-1"))
			})

			It("should set verbose to false by default", func() {
				cfg, err := config.Load(configPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Verbose).To(BeFalse())
			})
		})

		Context("when config file does not exist", func() {
			It("should return an error", func() {
				_, err := config.Load("/nonexistent/config.yaml")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read config file"))
			})
		})

		Context("when config file has invalid YAML", func() {
			BeforeEach(func() {
				invalidYAML := `
hub:
  kubeconfig: /path
  invalid yaml here: [unclosed
`
				err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a parse error", func() {
				_, err := config.Load(configPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse config"))
			})
		})

		Context("when config has missing required fields", func() {
			BeforeEach(func() {
				incompleteConfig := `
verbose: true
`
				err := os.WriteFile(configPath, []byte(incompleteConfig), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a validation error", func() {
				_, err := config.Load(configPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})
	})

	Describe("Config Validation", func() {
		// Table-driven test within BDD structure
		DescribeTable("validating hub configuration",
			func(hubConfig config.HubConfig, expectedError string) {
				cfg := &config.Config{
					Hub: hubConfig,
					Defaults: config.Defaults{
						Spoke: config.SpokeDefaults{
							Provider: "aws",
						},
					},
				}

				err := cfg.Validate()
				if expectedError == "" {
					Expect(err).NotTo(HaveOccurred())
				} else {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(expectedError))
				}
			},
			Entry("valid hub config",
				config.HubConfig{
					Kubeconfig: "/path/to/kubeconfig",
					Context:    "hub-cluster",
					Namespace:  "open-cluster-management",
				},
				"",
			),
			Entry("missing kubeconfig",
				config.HubConfig{
					Context:   "hub-cluster",
					Namespace: "open-cluster-management",
				},
				"kubeconfig is required",
			),
			Entry("missing namespace",
				config.HubConfig{
					Kubeconfig: "/path/to/kubeconfig",
					Context:    "hub-cluster",
				},
				"namespace is required",
			),
		)
	})

	Describe("GetHubKubeconfig", func() {
		It("should return the hub kubeconfig path", func() {
			cfg := &config.Config{
				Hub: config.HubConfig{
					Kubeconfig: "/custom/path/kubeconfig",
					Context:    "test",
					Namespace:  "default",
				},
			}

			Expect(cfg.GetHubKubeconfig()).To(Equal("/custom/path/kubeconfig"))
		})
	})

	Describe("Default Configuration", func() {
		Context("when loading defaults", func() {
			It("should provide sensible defaults for missing optional fields", func() {
				cfg := config.NewDefaultConfig()

				Expect(cfg.Verbose).To(BeFalse())
				Expect(cfg.Hub.Namespace).To(Equal("open-cluster-management"))
			})
		})
	})
})
