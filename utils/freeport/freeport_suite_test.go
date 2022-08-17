package freeport_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFreeport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Free Port Suite")
}
