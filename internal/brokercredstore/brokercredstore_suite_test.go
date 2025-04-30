package brokercredstore_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrokercredstore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brokercredstore Suite")
}
