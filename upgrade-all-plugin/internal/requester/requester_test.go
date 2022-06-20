package requester_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requester", func() {

	var (
		fakeRequester requester.Requester
		fakeServer    *httptest.Server
		testReceiver  map[string]interface{}
	)

	Describe("NewRequester", func() {
		It("returns a requester with given values", func() {
			actualRequester := requester.NewRequester("test-url", "test-token", false)

			Expect(actualRequester.APIBaseURL).To(Equal("test-url"))
			Expect(actualRequester.APIToken).To(Equal("test-token"))
		})
	})

	FDescribe("Get", func() {
		BeforeEach(func() {
			testReceiver = map[string]interface{}{}

			fakeServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"test_value":"foo"}`))
			}))
			fakeRequester = requester.NewRequester(fakeServer.URL, "test-token", false)
		})

		It("fails if receiver is not pointer type", func() {
			err := fakeRequester.Get("", testReceiver)

			Expect(err).To(MatchError("reciever must be of type Pointer"))
		})

		//It("", func() {
		//
		//	expectedResponse := map[string]interface{}{"test_value": "foo"}
		//
		//	fakeRequester.Get("", &testReceiver)
		//
		//	Expect(testReceiver).To(Equal(expectedResponse))
		//})

		AfterEach(func() {
			fakeServer.Close()
		})
	})

})
