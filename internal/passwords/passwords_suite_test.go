package passwords_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPassword(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Suite")
}
