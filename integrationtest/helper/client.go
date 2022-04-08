package helper

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
	"github.com/onsi/gomega"
)

func (h *TestHelper) Client() *client.Client {
	brokerClient, err := client.New(h.username, h.password, "localhost", h.Port)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return brokerClient
}
