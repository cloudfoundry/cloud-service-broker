package infohandler_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/infohandler"
)

var _ = Describe("Info Handler", func() {
	var (
		server *httptest.Server
		client *http.Client
	)

	BeforeEach(func() {
		server = httptest.NewServer(infohandler.New(infohandler.Config{
			BrokerVersion: "fake-version",
			Uptime:        func() time.Duration { return time.Hour },
		}))
		client = server.Client()
	})

	AfterEach(func() {
		server.Close()
	})

	It("returns the correct payload", func() {
		resp, err := client.Get(server.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))
		Expect(resp).To(HaveHTTPBody(MatchJSON(`{"version":"fake-version","uptime":"1h0m0s"}`)))
	})
})
