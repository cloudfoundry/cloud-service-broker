package brokerpak_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrokerpak(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brokerpak Suite")
}
