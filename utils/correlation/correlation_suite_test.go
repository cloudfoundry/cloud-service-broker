package correlation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCorrelation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Correlation Suite")
}
