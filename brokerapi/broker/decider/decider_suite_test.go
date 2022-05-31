package decider_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDecider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Decider Suite")
}
