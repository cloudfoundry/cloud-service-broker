package requester_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Requester", func() {

	var (
		fakeRequester requester.Requester
		fakeServer    *ghttp.Server
		testReceiver  map[string]interface{}
	)

	Describe("NewRequester", func() {
		It("returns a requester with given values", func() {
			actualRequester := requester.NewRequester("test-url", "test-token", false)

			Expect(actualRequester.APIBaseURL).To(Equal("test-url"))
			Expect(actualRequester.APIToken).To(Equal("test-token"))
		})
	})

	Describe("Get", func() {
		BeforeEach(func() {
			testReceiver = map[string]interface{}{}

			fakeServer = ghttp.NewServer()

			fakeRequester = requester.NewRequester(fakeServer.URL(), "test-token", false)
		})

		When("request is valid", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/test-endpoint", ""),
						ghttp.RespondWith(http.StatusOK, `{"test_value": "foo"}`, nil),
					),
				)
			})
			It("fails if receiver is not pointer type", func() {
				err := fakeRequester.Get("test-endpoint", &testReceiver)

				Expect(err).NotTo(HaveOccurred())
				Expect(testReceiver).To(Equal(map[string]interface{}{"test_value": "foo"}))
			})
		})

		It("errors if receiver is not of type pointer", func() {
			err := fakeRequester.Get("test-endpoint", testReceiver)
			Expect(err).To(MatchError("receiver must be of type Pointer"))
		})

		When("request is invalid", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/not-a-real-url", ""),
						ghttp.RespondWith(http.StatusNotFound, "", nil),
					),
				)
			})

			It("returns an error", func() {
				err := fakeRequester.Get("not-a-real-url", &testReceiver)

				Expect(err).To(MatchError("http response: 404"))
			})
		})

		When("unable to parse response body as JSON", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/test-endpoint", ""),
						ghttp.RespondWith(http.StatusOK, ``, nil),
					),
				)
			})

			It("returns an error", func() {
				err := fakeRequester.Get("test-endpoint", &testReceiver)

				Expect(err).To(MatchError("failed to unmarshal response into receiver error: unexpected end of JSON input"))
			})
		})

		AfterEach(func() {
			fakeServer.Close()
		})
	})

})
