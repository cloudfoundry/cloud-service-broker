package paramparser_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestParamparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Paramparser Suite")
}
