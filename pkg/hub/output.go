package hub

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	// OutputFormatTable represents table output format
	OutputFormatTable OutputFormat = "table"
	// OutputFormatJSON represents JSON output format
	OutputFormatJSON OutputFormat = "json"
)

// OutputWriter handles formatting and writing cluster information
type OutputWriter struct {
	format OutputFormat
	writer io.Writer
}

// NewOutputWriter creates a new OutputWriter with the specified format and writer
func NewOutputWriter(format OutputFormat, writer io.Writer) *OutputWriter {
	return &OutputWriter{
		format: format,
		writer: writer,
	}
}

// Write formats and writes the cluster information according to the configured format
func (o *OutputWriter) Write(clusters []ManagedClusterInfo) error {
	switch o.format {
	case OutputFormatTable:
		return o.writeTable(clusters)
	case OutputFormatJSON:
		return o.writeJSON(clusters)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

// writeTable writes cluster information in table format
func (o *OutputWriter) writeTable(clusters []ManagedClusterInfo) error {
	// Create tabwriter for column alignment
	w := tabwriter.NewWriter(o.writer, 0, 0, 3, ' ', 0)

	// Write header
	fmt.Fprintf(w, "NAME\tSTATUS\tAVAILABLE\n")

	// Write cluster rows
	for _, cluster := range clusters {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			cluster.Name,
			cluster.Status,
			cluster.Available,
		)
	}

	// Flush the tabwriter to ensure all data is written
	return w.Flush()
}

// writeJSON writes cluster information in JSON format
func (o *OutputWriter) writeJSON(clusters []ManagedClusterInfo) error {
	// Use MarshalIndent for pretty-printed JSON with 2-space indentation
	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal clusters to JSON: %w", err)
	}

	// Write JSON to output
	_, err = o.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}

	// Add newline at the end
	_, err = o.writer.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}
