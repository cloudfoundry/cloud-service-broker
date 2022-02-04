package helper

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	"github.com/onsi/gomega"
)

func (tl *TestLab) Client() *client.Client {
	brokerClient, err := client.New(tl.username, tl.password, "localhost", tl.port)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return brokerClient
}
