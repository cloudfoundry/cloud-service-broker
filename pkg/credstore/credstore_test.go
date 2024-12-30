// Copyright 2020 Pivotal Software, Inc.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//    http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package credstore_test

import (
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/credstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credhub Store", func() {
	var logger lager.Logger

	var uaaTestServer *httptest.Server
	var uaaRequest *http.Request

	var chTestServer *httptest.Server
	var chRequest *http.Request

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		uaaTestServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{}`))
			uaaRequest = r
		}))
		chTestServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"data": [{"foo": "bar"}]}`))
			chRequest = r
		}))
	})

	It("sanity check (class under test is mostly just a wrapper)", func() {
		credStoreConfig := &config.CredStoreConfig{
			CredHubURL:           chTestServer.URL,
			UaaURL:               uaaTestServer.URL,
			UaaClientName:        "my-client",
			UaaClientSecret:      "my-secret",
			SkipSSLValidation:    true,
			CACert:               "",
			StoreBindCredentials: false,
		}
		chStore, err := credstore.NewCredhubStore(credStoreConfig, logger)
		Expect(err).To(BeNil())
		cred, err := chStore.Get("/foo/bar/baz")

		Expect(err).To(BeNil())
		Expect(cred).NotTo(BeNil())

		Expect(uaaRequest).NotTo(BeNil())

		Expect(chRequest).NotTo(BeNil())
		chRequest.ParseForm()
		Expect(chRequest.Form["name"]).To(Equal([]string{"/foo/bar/baz"}))
	})

	It("sanity check cert file (class under test is mostly just a wrapper", func() {
		content := `-----BEGIN CERTIFICATE-----
MIICsDCCAZgCCQDi7u3xz4OO2TANBgkqhkiG9w0BAQsFADAaMRgwFgYDVQQDDA93
d3cuZXhhbXBsZS5jb20wHhcNMTkxMDI0MTkyNzEyWhcNMjkxMDIxMTkyNzEyWjAa
MRgwFgYDVQQDDA93d3cuZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQD0eWkUfSWvCq2s21+rPPiEnxVp8WLPnTvz1o8EOShKHYv5TRf5
oRC7jHVw+fzcJQYZ4bImjlSaezGVMutUPod4l0lWsZBeIQHLVVO4dIWc3U/CECJA
pfK4EmUbbLquDILYLX+GqXgXPdBNm9FubRiAIAolInZBaXlKv1AO49IvVL3lrXWa
LbtY7FqaSAuZEMRNBMLdehSIbpKXHLzXUw5+RzQ2jIy6qS/UpV/9SSoooy+AcqRb
5Lo48myNe8ozT5AsxEw4/o4mh5vfg2j4Zwt/2h7LbMedXKUpgO8Dhlkt1vHWGSZH
bGPRcvVZ5va9RPYzrG4xyvCg1UjLDPAjy+YjAgMBAAEwDQYJKoZIhvcNAQELBQAD
ggEBABuJfMbYpQfMHizsxhEa3U6mdIGcRh93U9LtGYMsQslRbY5/Bz7KgkI3+xUh
fHaWix3GA2HipNfpNtbIxvrj5lrSsNw5vl39TsDXEwbC8wgqKWQCi+8cilIuDEpS
WiaQqAkK41aqRSDOzUV4worM5HEeFGmSowrLRJOk1Wf1EGw8fD51pO3Zl4hv+PxN
/hSSD7b12tEf59WnppMDXEvXbVVUbc1bQKrUBdbqRRAIvdVXkQQVKd11JAcsTi/T
DLgHJRgZ5Bkp6yhm2RRzuQeMbozry9wXJg9MN14aLjfUNkB08+BX7kDk4H7ZgQ4S
GqdyaKP2/eZ04RHn1TYI/UGRnzk=
-----END CERTIFICATE-----`

		credStoreConfig := &config.CredStoreConfig{
			CredHubURL:           chTestServer.URL,
			UaaURL:               uaaTestServer.URL,
			UaaClientName:        "my-client",
			UaaClientSecret:      "my-secret",
			SkipSSLValidation:    true,
			CACert:               content,
			StoreBindCredentials: false,
		}
		chStore, err := credstore.NewCredhubStore(credStoreConfig, logger)

		Expect(err).To(BeNil())
		cred, err := chStore.Get("/foo/bar/baz")

		Expect(err).To(BeNil())
		Expect(cred).NotTo(BeNil())

		Expect(uaaRequest).NotTo(BeNil())

		Expect(chRequest).NotTo(BeNil())
		chRequest.ParseForm()
		Expect(chRequest.Form["name"]).To(Equal([]string{"/foo/bar/baz"}))
	})
})
