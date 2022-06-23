package upgrader_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpgrader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrader Suite")
}
