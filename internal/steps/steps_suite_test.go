package steps_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSteps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Steps Suite")
}
