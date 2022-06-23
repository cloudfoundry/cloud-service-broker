package workers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workers Suite")
}
