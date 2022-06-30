package command_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader/upgraderfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeAll", func() {

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeLogger        *upgraderfakes.FakeLogger
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeLogger = &upgraderfakes.FakeLogger{}

		fakeCliConnection.ApiVersionReturns("3.99.0", nil)
		fakeCliConnection.IsLoggedInReturns(true, nil)
		fakeCliConnection.AccessTokenReturns("access-token", nil)
		fakeCliConnection.ApiEndpointReturns("test-endpoint", nil)
		fakeCliConnection.IsSSLDisabledReturns(false, nil)
	})

	When("invalid input is given", func() {
		It("returns an error", func() {
			err := command.UpgradeAll(fakeCliConnection, []string{}, fakeLogger)

			Expect(err).To(MatchError(fmt.Errorf("broker name must be specifed\nusage:\ncf upgrade-all-service-instances <broker-name>")))
		})
	})

	When("access token can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.AccessTokenReturns("", fmt.Errorf("AccessToken error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"}, fakeLogger)
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api access token: AccessToken error")))
		})
	})

	When("api endpoint can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.ApiEndpointReturns("", fmt.Errorf("APIEndpoint error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"}, fakeLogger)
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api endpoint: APIEndpoint error")))
		})
	})

	When("ssl validation can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.IsSSLDisabledReturns(false, fmt.Errorf("ssl error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"}, fakeLogger)
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api ssl validation status: ssl error")))
		})
	})

	When("upgrade errors", func() {
		It("returns the error", func() {
			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"}, fakeLogger)
			Expect(err).To(HaveOccurred())
		})
	})

})
