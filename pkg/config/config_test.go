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

package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/cloud-service-broker/pkg/config"
)

var _ = Describe("Config", func() {
	BeforeEach(func() {
		os.Clearenv()
	})

	Context("config parsing", func() {
		It("sets defaults", func() {
			c, err := Parse()
			Expect(err).To(BeNil())
			Expect(c).ToNot(BeNil())

			Expect(c.CredStoreConfig.HasCredHubConfig()).To(BeFalse())
		})

		Context("credstore config", func() {
			It("parses credstore config", func() {
				os.Setenv("CH_CRED_HUB_URL", "https://credhub.example.com")
				os.Setenv("CH_UAA_URL", "https://uaa.example.com")
				os.Setenv("CH_UAA_CLIENT_NAME", "my-client")
				os.Setenv("CH_UAA_CLIENT_SECRET", "my-secret")
				os.Setenv("CH_SKIP_SSL_VALIDATION", "true")

				c, err := Parse()
				Expect(err).To(BeNil())
				Expect(c).ToNot(BeNil())

				Expect(c.CredStoreConfig.HasCredHubConfig()).To(BeTrue())

				Expect(c.CredStoreConfig.CredHubURL).To(Equal("https://credhub.example.com"))
				Expect(c.CredStoreConfig.UaaURL).To(Equal("https://uaa.example.com"))
			})
		})
	})
})
