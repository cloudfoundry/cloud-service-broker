package stableuuid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStableUUID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StableUUID Suite")
}
