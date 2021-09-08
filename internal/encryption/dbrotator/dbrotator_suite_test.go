package dbrotator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDBRotator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DB Encryption Rotation Suite")
}
