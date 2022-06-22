package command_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mi struct {
	Version string `json:"version"`
}

var _ = Describe("UpgradeAll", func() {

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		//fakeServer        *ghttp.Server
		//responseCode      int
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
	})

	When("invalid input is given", func() {
		It("returns an error", func() {
			err := command.UpgradeAll(fakeCliConnection, []string{})

			Expect(err).To(MatchError(fmt.Errorf("broker name must be specifed")))
		})
	})

	When("access token can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
			fakeCliConnection.IsLoggedInReturns(true, nil)
			fakeCliConnection.AccessTokenReturns("", fmt.Errorf("AccessToken error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"})
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api access token: AccessToken error")))
		})
	})

	When("api endpoint can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
			fakeCliConnection.IsLoggedInReturns(true, nil)
			fakeCliConnection.AccessTokenReturns("access-token", nil)
			fakeCliConnection.ApiEndpointReturns("", fmt.Errorf("APIEndpoint error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"})
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api endpoint: APIEndpoint error")))
		})
	})

	// validate input
	// gets token and endpoint
	// parses flags

})
