package hub_test

import (
	"bytes"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
)

var _ = Describe("OutputWriter", func() {
	var (
		buffer   *bytes.Buffer
		writer   *hub.OutputWriter
		clusters []hub.ManagedClusterInfo
	)

	BeforeEach(func() {
		buffer = new(bytes.Buffer)
		clusters = []hub.ManagedClusterInfo{
			{Name: "cluster-east-1", Status: hub.StatusReady, Available: "True", Message: "Cluster is healthy"},
			{Name: "cluster-west-1", Status: hub.StatusNotReady, Available: "False", Message: "Cluster is not available"},
			{Name: "cluster-central", Status: hub.StatusUnknown, Available: "Unknown", Message: ""},
		}
	})

	Describe("Table Output", func() {
		BeforeEach(func() {
			writer = hub.NewOutputWriter(hub.OutputFormatTable, buffer)
		})

		Context("with multiple clusters", func() {
			It("should format output as a table with headers", func() {
				err := writer.Write(clusters)
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()
				lines := strings.Split(strings.TrimSpace(output), "\n")

				// Check header
				Expect(lines[0]).To(ContainSubstring("NAME"))
				Expect(lines[0]).To(ContainSubstring("STATUS"))
				Expect(lines[0]).To(ContainSubstring("AVAILABLE"))

				// Check that all clusters are present
				Expect(output).To(ContainSubstring("cluster-east-1"))
				Expect(output).To(ContainSubstring("cluster-west-1"))
				Expect(output).To(ContainSubstring("cluster-central"))

				// Check status values
				Expect(output).To(ContainSubstring("Ready"))
				Expect(output).To(ContainSubstring("NotReady"))
				Expect(output).To(ContainSubstring("Unknown"))

				// Check available values
				Expect(output).To(ContainSubstring("True"))
				Expect(output).To(ContainSubstring("False"))
			})

			It("should align columns properly", func() {
				err := writer.Write(clusters)
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()
				lines := strings.Split(strings.TrimSpace(output), "\n")

				// Should have header + 3 cluster rows
				Expect(lines).To(HaveLen(4))

				// Each line should have content (not just whitespace)
				for _, line := range lines {
					Expect(strings.TrimSpace(line)).NotTo(BeEmpty())
				}
			})
		})

		Context("with empty cluster list", func() {
			It("should display only headers", func() {
				err := writer.Write([]hub.ManagedClusterInfo{})
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()
				lines := strings.Split(strings.TrimSpace(output), "\n")

				// Should only have the header line
				Expect(lines).To(HaveLen(1))
				Expect(lines[0]).To(ContainSubstring("NAME"))
				Expect(lines[0]).To(ContainSubstring("STATUS"))
				Expect(lines[0]).To(ContainSubstring("AVAILABLE"))
			})
		})

		Context("with single cluster", func() {
			It("should format correctly", func() {
				singleCluster := []hub.ManagedClusterInfo{
					{Name: "my-cluster", Status: hub.StatusReady, Available: "True"},
				}

				err := writer.Write(singleCluster)
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()
				Expect(output).To(ContainSubstring("NAME"))
				Expect(output).To(ContainSubstring("my-cluster"))
				Expect(output).To(ContainSubstring("Ready"))
				Expect(output).To(ContainSubstring("True"))
			})
		})
	})

	Describe("JSON Output", func() {
		BeforeEach(func() {
			writer = hub.NewOutputWriter(hub.OutputFormatJSON, buffer)
		})

		Context("with multiple clusters", func() {
			It("should format output as valid JSON", func() {
				err := writer.Write(clusters)
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()

				// Verify it's valid JSON
				var result []hub.ManagedClusterInfo
				err = json.Unmarshal([]byte(output), &result)
				Expect(err).NotTo(HaveOccurred())

				// Verify all clusters are present
				Expect(result).To(HaveLen(3))
			})

			It("should preserve cluster data accurately", func() {
				err := writer.Write(clusters)
				Expect(err).NotTo(HaveOccurred())

				var result []hub.ManagedClusterInfo
				err = json.Unmarshal(buffer.Bytes(), &result)
				Expect(err).NotTo(HaveOccurred())

				// Verify cluster data
				clusterMap := make(map[string]hub.ManagedClusterInfo)
				for _, c := range result {
					clusterMap[c.Name] = c
				}

				Expect(clusterMap["cluster-east-1"].Status).To(Equal(hub.StatusReady))
				Expect(clusterMap["cluster-east-1"].Available).To(Equal("True"))
				Expect(clusterMap["cluster-east-1"].Message).To(Equal("Cluster is healthy"))

				Expect(clusterMap["cluster-west-1"].Status).To(Equal(hub.StatusNotReady))
				Expect(clusterMap["cluster-west-1"].Available).To(Equal("False"))

				Expect(clusterMap["cluster-central"].Status).To(Equal(hub.StatusUnknown))
				Expect(clusterMap["cluster-central"].Available).To(Equal("Unknown"))
			})

			It("should be pretty-printed with indentation", func() {
				err := writer.Write(clusters)
				Expect(err).NotTo(HaveOccurred())

				output := buffer.String()

				// Check for indentation (pretty printing)
				Expect(output).To(ContainSubstring("  "))

				// Should be multiple lines (not minified)
				lines := strings.Split(output, "\n")
				Expect(len(lines)).To(BeNumerically(">", 5))
			})
		})

		Context("with empty cluster list", func() {
			It("should return empty JSON array", func() {
				err := writer.Write([]hub.ManagedClusterInfo{})
				Expect(err).NotTo(HaveOccurred())

				output := strings.TrimSpace(buffer.String())

				// Verify it's an empty JSON array
				var result []hub.ManagedClusterInfo
				err = json.Unmarshal([]byte(output), &result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEmpty())

				// Should be "[]" (possibly with newline)
				Expect(output).To(ContainSubstring("[]"))
			})
		})

		Context("with single cluster", func() {
			It("should format correctly", func() {
				singleCluster := []hub.ManagedClusterInfo{
					{Name: "my-cluster", Status: hub.StatusReady, Available: "True", Message: "All good"},
				}

				err := writer.Write(singleCluster)
				Expect(err).NotTo(HaveOccurred())

				var result []hub.ManagedClusterInfo
				err = json.Unmarshal(buffer.Bytes(), &result)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(HaveLen(1))
				Expect(result[0].Name).To(Equal("my-cluster"))
				Expect(result[0].Status).To(Equal(hub.StatusReady))
				Expect(result[0].Available).To(Equal("True"))
				Expect(result[0].Message).To(Equal("All good"))
			})
		})
	})

	Describe("NewOutputWriter", func() {
		It("should create a writer with table format", func() {
			writer := hub.NewOutputWriter(hub.OutputFormatTable, buffer)
			Expect(writer).NotTo(BeNil())
		})

		It("should create a writer with JSON format", func() {
			writer := hub.NewOutputWriter(hub.OutputFormatJSON, buffer)
			Expect(writer).NotTo(BeNil())
		})
	})

	Describe("Error Handling", func() {
		It("should return error for unsupported output format", func() {
			writer := hub.NewOutputWriter(hub.OutputFormat("invalid"), buffer)
			err := writer.Write(clusters)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported output format"))
		})
	})
})
