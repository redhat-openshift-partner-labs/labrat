//go:build test

package spoke_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpoke(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spoke Suite")
}
