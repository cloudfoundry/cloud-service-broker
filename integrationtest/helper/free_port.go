package helper

import (
	"net"

	"github.com/onsi/gomega"
)

func freePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
