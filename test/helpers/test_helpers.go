package helpers

import (
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

// CreateTempConfigFile creates a temporary config file with the given content
// and returns the path. The caller is responsible for cleanup.
func CreateTempConfigFile(content string) string {
	tempDir, err := os.MkdirTemp("", "labrat-test-")
	Expect(err).NotTo(HaveOccurred())

	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(content), 0644)
	Expect(err).NotTo(HaveOccurred())

	return configPath
}

// CleanupTempDir removes a temporary directory
func CleanupTempDir(path string) {
	if path != "" {
		os.RemoveAll(filepath.Dir(path))
	}
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
