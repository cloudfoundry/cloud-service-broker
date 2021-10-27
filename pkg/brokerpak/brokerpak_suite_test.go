package brokerpak_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBrokerpak(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brokerpak Suite")
}
