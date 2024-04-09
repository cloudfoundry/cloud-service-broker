package httpsmiddleware_test

import (
	"net/http"
	"net/url"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/cloudfoundry/cloud-service-broker/internal/httpsmiddleware"
	"github.com/cloudfoundry/cloud-service-broker/internal/httpsmiddleware/httpsmiddlewarefakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate net/http.Handler
//counterfeiter:generate net/http.ResponseWriter

var _ = Describe("EnsureHTTPS()", func() {
	const xForwardProto = "X-Forwarded-Proto"

	var (
		logger             *lagertest.TestLogger
		fakeNext           *httpsmiddlewarefakes.FakeHandler
		fakeResponseWriter *httpsmiddlewarefakes.FakeResponseWriter
		fakeResponseHeader http.Header
		fakeURL            url.URL
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		fakeNext = &httpsmiddlewarefakes.FakeHandler{}
		fakeURL = url.URL{
			Scheme: "http",
			Host:   "fake-host",
			Path:   "fake-path",
		}

		fakeResponseWriter = &httpsmiddlewarefakes.FakeResponseWriter{}
		fakeResponseHeader = make(http.Header)
		fakeResponseWriter.HeaderReturns(fakeResponseHeader)
	})

	It("does nothing when disabled", func() {
		httpsmiddleware.EnsureHTTPS(fakeNext, logger, true).ServeHTTP(fakeResponseWriter, &http.Request{
			Method: http.MethodHead,
			URL:    &fakeURL,
			Header: http.Header{xForwardProto: []string{"http"}},
		})

		Expect(fakeNext.ServeHTTPCallCount()).To(Equal(1))
		Expect(fakeResponseWriter.WriteHeaderCallCount()).To(BeZero())
	})

	DescribeTable(
		"redirects non-HTTPS connections",
		func(header http.Header) {
			httpsmiddleware.EnsureHTTPS(fakeNext, logger, false).ServeHTTP(fakeResponseWriter, &http.Request{
				Method: http.MethodHead,
				URL:    &fakeURL,
				Header: header,
			})

			Expect(fakeNext.ServeHTTPCallCount()).To(BeZero())
			Expect(fakeResponseWriter.WriteHeaderCallCount()).To(Equal(1))
			Expect(fakeResponseWriter.WriteHeaderArgsForCall(0)).To(Equal(http.StatusMovedPermanently))
			Expect(fakeResponseHeader.Get("Location")).To(Equal("https://http://fake-host/fake-path"))

			Expect(logger.Buffer().Contents()).To(ContainSubstring("redirecting-to-https"))
		},
		Entry("empty", nil),
		Entry("http", http.Header{xForwardProto: []string{"http"}}),
		Entry("other", http.Header{xForwardProto: []string{"foo"}}),
		Entry("https with extra (comma separated)", http.Header{xForwardProto: []string{"https,foo"}}),
		Entry("https with extra (slice)", http.Header{xForwardProto: []string{"https", "foo"}}),
	)

	It("allows an HTTPS connection", func() {
		httpsmiddleware.EnsureHTTPS(fakeNext, logger, false).ServeHTTP(fakeResponseWriter, &http.Request{
			Method: http.MethodHead,
			URL:    &fakeURL,
			Header: http.Header{xForwardProto: []string{"https"}},
		})

		Expect(fakeNext.ServeHTTPCallCount()).To(Equal(1))
		Expect(fakeResponseWriter.WriteHeaderCallCount()).To(BeZero())
	})
})
