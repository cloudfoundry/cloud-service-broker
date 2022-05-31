package decider_test

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/decider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
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
		serviceDefinition *broker.ServiceDefinition

		defaultMI *domain.MaintenanceInfo
		higherMI  *domain.MaintenanceInfo
	)

	BeforeEach(func() {
		defaultMI = &domain.MaintenanceInfo{
			Version: "1.2.3",
		}

		higherMI = &domain.MaintenanceInfo{
			Version: "1.2.4",
		}

		serviceDefinition = &broker.ServiceDefinition{
			ID: "fake-service-id",
			Plans: []broker.ServicePlan{
				{
					ServicePlan: domain.ServicePlan{ID: planWithMI, MaintenanceInfo: defaultMI},
				},
				{
					ServicePlan: domain.ServicePlan{ID: otherPlanWithMI, MaintenanceInfo: higherMI},
				},
				{
					ServicePlan: domain.ServicePlan{ID: planWithoutMI},
				},
				{
					ServicePlan: domain.ServicePlan{ID: otherPlanWithoutMI},
				},
			},
		}
	})

	Describe("DecideOperation", func() {
		It("fails when the requested plan is not in the catalog", func() {
			details := domain.UpdateDetails{
				PlanID: "not-in-catalog",
			}

			_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
			Expect(err).To(MatchError("plan not-in-catalog does not exist"))
		})

		Context("request without maintenance_info - user is attempting update", func() {
			It("does not error when the catalog's plan doesn't have maintenance info either", func() {
				details := domain.UpdateDetails{
					PlanID: otherPlanWithoutMI,
					PreviousValues: domain.PreviousValues{
						PlanID: planWithoutMI,
					},
				}

				_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
				Expect(err).NotTo(HaveOccurred())
			})

			When("the request is a change of plan", func() {
				It("is an update", func() {
					details := domain.UpdateDetails{
						PlanID: otherPlanWithoutMI,
						PreviousValues: domain.PreviousValues{
							PlanID: planWithoutMI,
						},
					}

					operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("there are request parameters", func() {
				It("is an update", func() {
					details := domain.UpdateDetails{
						PlanID:        planWithoutMI,
						RawParameters: json.RawMessage(`{"foo": "bar"}`),
					}

					operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("the plan does not change and there are no request parameters", func() {
				It("is an update", func() {
					details := domain.UpdateDetails{
						PlanID:        planWithoutMI,
						RawParameters: json.RawMessage(`{ }`),
						PreviousValues: domain.PreviousValues{
							PlanID: planWithoutMI,
						},
					}

					operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
					Expect(err).NotTo(HaveOccurred())
					Expect(operation).To(Equal(decider.Update))
				})
			})

			When("the desired plan has maintenance_info in the catalog", func() {
				When("no previous maintenance_info is present in the request", func() {
					When("the previous plan has maintenance info", func() {
						It("should fail", func() {
							details := domain.UpdateDetails{
								PlanID: otherPlanWithMI,
								PreviousValues: domain.PreviousValues{
									PlanID: planWithMI,
								},
							}

							operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
							Expect(err).To(MatchError("service instance needs to be upgraded before updating: maintenance info defined in broker service catalog, but not passed in request"))
							Expect(operation).To(Equal(decider.Failed))
						})
					})
				})

				When("previous maintenance_info is present in the request", func() {
					It("fails when it does not match the catalog's maintenance info for the previous plan", func() {
						details := domain.UpdateDetails{
							PlanID: otherPlanWithMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: higherMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).To(MatchError("service instance needs to be upgraded before updating: maintenance info defined in broker service catalog, but not passed in request"))
						Expect(operation).To(Equal(decider.Failed))
					})

					It("fails when it matches the catalog's maintenance info for the previous plan", func() {
						details := domain.UpdateDetails{
							PlanID: otherPlanWithMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: defaultMI,
							},
						}

						op, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).To(MatchError("service instance needs to be upgraded before updating: maintenance info defined in broker service catalog, but not passed in request"))
						Expect(op).To(Equal(decider.Failed))
					})
				})
			})
		})

		Context("request with maintenance_info", func() {
			Context("request and plan have the same maintenance_info", func() {
				When("the request is a change of plan", func() {
					It("is an update", func() {
						details := domain.UpdateDetails{
							PlanID:          otherPlanWithMI,
							MaintenanceInfo: higherMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: defaultMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

				When("there are request parameters", func() {
					It("is an update", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							RawParameters:   json.RawMessage(`{"foo": "bar"}`),
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

				When("the plan does not change and there are no request parameters", func() {
					It("is an update", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							RawParameters:   json.RawMessage(`{ }`),
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: defaultMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})
				})

			})

			Context("request has different maintenance_info values", func() {
				When("adding maintenance_info when there was none before", func() {
					It("is an upgrade", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							PreviousValues: domain.PreviousValues{
								PlanID: planWithMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("removing maintenance_info when it was there before", func() {
					It("is an upgrade", func() {
						details := domain.UpdateDetails{
							PlanID: planWithoutMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithoutMI,
								MaintenanceInfo: defaultMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("the plan has not changed and there are no request parameters", func() {
					It("is an upgrade", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: higherMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Upgrade))
					})
				})

				When("there is a change of plan", func() {
					It("fails when the previous maintenance_info does not match the previous plan maintenance_info", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          otherPlanWithMI,
								MaintenanceInfo: defaultMI,
							},
						}

						_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).To(MatchError(apiresponses.NewFailureResponseBuilder(
							errors.New("service instance needs to be upgraded before updating"),
							http.StatusUnprocessableEntity,
							"previous-maintenance-info-check",
						).Build()))
					})

					It("is an update when the previous maintenance_info matches the previous plan", func() {
						details := domain.UpdateDetails{
							PlanID:          otherPlanWithMI,
							MaintenanceInfo: higherMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: defaultMI,
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).NotTo(HaveOccurred())
						Expect(operation).To(Equal(decider.Update))
					})

					It("is an update when the previous plan is not in the catalog", func() {
						details := domain.UpdateDetails{
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							PreviousValues: domain.PreviousValues{
								PlanID: "fake-plan-that-does-not-exist",
								MaintenanceInfo: &domain.MaintenanceInfo{
									Version: "1.2.1",
								},
							},
						}

						operation, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
						Expect(err).To(MatchError("service instance needs to be upgraded: plan fake-plan-that-does-not-exist does not exist. Contact the operator for assistance"))
						Expect(operation).To(Equal(decider.Failed))
					})
				})

				When("there are request parameters", func() {
					It("fails", func() {
						details := domain.UpdateDetails{
							RawParameters:   json.RawMessage(`{"foo": "bar"}`),
							PlanID:          planWithMI,
							MaintenanceInfo: defaultMI,
							PreviousValues: domain.PreviousValues{
								PlanID:          planWithMI,
								MaintenanceInfo: higherMI,
							},
						}

						_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
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
					details := domain.UpdateDetails{
						PlanID:          planWithMI,
						MaintenanceInfo: higherMI,
					}

					_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)

					Expect(err).To(MatchError(apiresponses.ErrMaintenanceInfoConflict))
				})

				It("fails when the request has maintenance_info but the plan does not", func() {
					details := domain.UpdateDetails{
						PlanID:          planWithoutMI,
						MaintenanceInfo: defaultMI,
					}

					_, err := decider.Decider{}.DecideOperation(serviceDefinition, details)
					Expect(err).To(MatchError(apiresponses.ErrMaintenanceInfoNilConflict))
				})
			})
		})
	})
})
