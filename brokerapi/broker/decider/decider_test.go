package decider_test

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/decider"
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

var _ = Describe("Decider", func() {
	const (
		planWithMI         = "fake-plan-id-with-mi"
		otherPlanWithMI    = "fake-other-plan-id-with-mi"
		planWithoutMI      = "fake-plan-id-no-mi"
		otherPlanWithoutMI = "fake-other-plan-id-no-mi"
	)
	var (
		defaultMI *version.Version
		higherMI  *version.Version
	)

	BeforeEach(func() {
		defaultMI = version.Must(version.NewVersion("1.2.3"))
		higherMI = version.Must(version.NewVersion("1.2.4"))
	})

	Describe("DecideOperation", func() {
		Context("request without maintenance_info", func() {
			It("does not error when the catalog's plan doesn't have maintenance info either", func() {
				details := paramparser.UpdateDetails{
					PlanID:         otherPlanWithoutMI,
					PreviousPlanID: planWithoutMI,
				}

				_, err := decider.DecideOperation(nil, details)
				Expect(err).NotTo(HaveOccurred())
			})

			When("the request is a change of plan", func() {
				It("is an update", func() {
					details := paramparser.UpdateDetails{
						PlanID:         otherPlanWithoutMI,
						PreviousPlanID: planWithoutMI,
					}

					operation, err := decider.DecideOperation(nil, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("there are request parameters", func() {
				It("is an update", func() {
					details := paramparser.UpdateDetails{
						PlanID:        planWithoutMI,
						RequestParams: map[string]any{"foo": "bar"},
					}

					operation, err := decider.DecideOperation(nil, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("the plan does not change and there are no request parameters", func() {
				It("is an update", func() {
					details := paramparser.UpdateDetails{
						PlanID:         planWithoutMI,
						RequestParams:  map[string]any{},
						PreviousPlanID: planWithoutMI,
					}

					operation, err := decider.DecideOperation(nil, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("the desired plan has maintenance_info in the catalog", func() {
				It("fails", func() {
					details := paramparser.UpdateDetails{
						PlanID: otherPlanWithMI,
					}

					operation, err := decider.DecideOperation(defaultMI, details)
					Expect(err).To(MatchError("service instance needs to be upgraded before updating: maintenance info defined in broker service catalog, but not passed in request"))
					Expect(operation).To(Equal(decider.Failed))
				})
			})
		})

		Context("request with maintenance_info", func() {
			Context("request and plan have the same maintenance_info", func() {
				When("the request is a change of plan", func() {
					It("is an update", func() {
						details := paramparser.UpdateDetails{
							PlanID:                         otherPlanWithMI,
							MaintenanceInfoVersion:         defaultMI,
							PreviousPlanID:                 planWithMI,
							PreviousMaintenanceInfoVersion: defaultMI,
						}

						operation, err := decider.DecideOperation(defaultMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

				When("there are request parameters", func() {
					It("is an update", func() {
						details := paramparser.UpdateDetails{
							PlanID:                 planWithMI,
							MaintenanceInfoVersion: defaultMI,
							RequestParams:          map[string]any{"foo": "bar"},
						}

						operation, err := decider.DecideOperation(defaultMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

				When("the plan does not change and there are no request parameters", func() {
					It("is an update", func() {
						details := paramparser.UpdateDetails{
							PlanID:                         planWithMI,
							MaintenanceInfoVersion:         defaultMI,
							RequestParams:                  map[string]any{},
							PreviousPlanID:                 planWithMI,
							PreviousMaintenanceInfoVersion: defaultMI,
						}

						operation, err := decider.DecideOperation(defaultMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})
			})

			Context("request has different maintenance_info values to the plan", func() {
				When("adding maintenance_info when there was none before", func() {
					It("is an upgrade", func() {
						details := paramparser.UpdateDetails{
							PlanID:                 planWithMI,
							MaintenanceInfoVersion: defaultMI,
							PreviousPlanID:         planWithMI,
						}

						operation, err := decider.DecideOperation(defaultMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("removing maintenance_info when it was there before", func() {
					It("is an upgrade", func() {
						details := paramparser.UpdateDetails{
							PlanID:                         planWithoutMI,
							PreviousPlanID:                 planWithoutMI,
							PreviousMaintenanceInfoVersion: defaultMI,
						}

						operation, err := decider.DecideOperation(nil, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("the plan has not changed and there are no request parameters", func() {
					It("is an upgrade", func() {
						details := paramparser.UpdateDetails{
							PlanID:                         planWithMI,
							MaintenanceInfoVersion:         defaultMI,
							PreviousPlanID:                 planWithMI,
							PreviousMaintenanceInfoVersion: higherMI,
						}

						operation, err := decider.DecideOperation(defaultMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("there is a change of plan", func() {
					It("is an update when the previous maintenance_info differs from the requested maintenance_info", func() {
						details := paramparser.UpdateDetails{
							PlanID:                         otherPlanWithMI,
							MaintenanceInfoVersion:         higherMI,
							PreviousPlanID:                 planWithMI,
							PreviousMaintenanceInfoVersion: higherMI,
						}

						operation, err := decider.DecideOperation(higherMI, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

				When("there are request parameters", func() {
					It("fails", func() {
						details := paramparser.UpdateDetails{
							RequestParams:                  map[string]any{"foo": "bar"},
							PlanID:                         planWithMI,
							MaintenanceInfoVersion:         defaultMI,
							PreviousPlanID:                 planWithMI,
							PreviousMaintenanceInfoVersion: higherMI,
						}

						_, err := decider.DecideOperation(defaultMI, details)
						Expect(err).To(MatchError(apiresponses.NewFailureResponseBuilder(
							errors.New("service instance needs to be upgraded before updating"),
							http.StatusUnprocessableEntity,
							"previous-maintenance-info-check",
						).Build()))
					})
				})
			})

			Context("request and plan have different maintenance_info values", func() {
				It("fails when the maintenance_info requested does not match the plan", func() {
					details := paramparser.UpdateDetails{
						PlanID:                 planWithMI,
						MaintenanceInfoVersion: higherMI,
					}

					_, err := decider.DecideOperation(defaultMI, details)

					Expect(err).To(MatchError(apiresponses.ErrMaintenanceInfoConflict))
				})

				It("fails when the request has maintenance_info but the plan does not", func() {
					details := paramparser.UpdateDetails{
						PlanID:                 planWithoutMI,
						MaintenanceInfoVersion: defaultMI,
					}

					_, err := decider.DecideOperation(nil, details)
					Expect(err).To(MatchError(apiresponses.ErrMaintenanceInfoNilConflict))
				})
			})
		})
	})
})
