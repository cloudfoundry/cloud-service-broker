package tf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTF(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TF Suite")
}
