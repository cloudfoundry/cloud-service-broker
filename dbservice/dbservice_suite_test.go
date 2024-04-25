package dbservice

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDBService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DBService Suite")
}
