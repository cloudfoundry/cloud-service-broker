package hclparser_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHclparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hclparser Suite")
}
