package validate_test

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/validate"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validate", func() {

	var (
		fakeCliConnection pluginfakes.FakeCliConnection
		fakeLogger        *log.Logger
	)

	BeforeEach(func() {
		fakeCliConnection = pluginfakes.FakeCliConnection{}
		fakeLogger = log.Default()
	})

	Describe("brokername validation", func() {
		When("no broker name is given", func() {
			It("fails to run the upgrade", func() {
				err := validate.ValidateInput(&fakeCliConnection, nil)

				Expect(err).To(MatchError(fmt.Errorf("broker name must be specifed")))
			})
		})

		When("invalid brokername is given", func() {
			It("returns an error", func() {
				err := validate.ValidateInput(&fakeCliConnection, []string{"*inValid'Broker/Name"})

				Expect(err).To(MatchError(fmt.Errorf("invalid brokername format")))
			})
		})
	})

	Describe("validateAPIVersion", func() {
		When("cf api version < 7", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("2.0.0", nil)

				err := validate.ValidateInput(&fakeCliConnection, []string{"broker-name"})

				Expect(err).To(MatchError("plugin requires CF API version >= 3.0.0"))
			})
		})
		When("unable to get API version", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("", fmt.Errorf("ApiVersion error"))

				err := validate.ValidateInput(&fakeCliConnection, []string{"broker-name"})

				Expect(err).To(MatchError("error retrieving api version: ApiVersion error"))
			})
		})
	})

	Describe("validateLoggedIn", func() {
		When("not authenticated", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("3.0.0", nil)
				fakeCliConnection.IsLoggedInReturns(false, nil)

				err := command.UpgradeAll(&fakeCliConnection, []string{"broker-name"}, fakeLogger)

				Expect(err).To(MatchError("you must authenticate with the cf cli before running this command"))
			})
		})
		When("unable to check if logged in", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("3.0.0", nil)
				fakeCliConnection.IsLoggedInReturns(false, fmt.Errorf("isLoggedIn error"))

				err := command.UpgradeAll(&fakeCliConnection, []string{"broker-name"}, fakeLogger)

				Expect(err).To(MatchError("error validating user authentication: isLoggedIn error"))
			})
		})
	})
})
