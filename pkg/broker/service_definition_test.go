package broker_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("ServiceDefinition", func() {
	Describe("Validate", func() {
		Context("validate ID", func() {
			It("should fail when Id is missing", func() {
				definition := broker.ServiceDefinition{Name: "test"}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: Id"))
			})

			It("should fail when Id is not valid", func() {
				definition := broker.ServiceDefinition{Id: "test", Name: "test-name"}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: Id"))
			})
		})

		It("should fail when Name is missing", func() {
			definition := broker.ServiceDefinition{Id: "55ad8194-0431-11ec-948a-63ff62e94b14"}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must match '^[a-zA-Z0-9-\\.]+$': Name"))
		})

		It("should fail when ImageUrl is not a URL", func() {
			definition := broker.ServiceDefinition{
				Id:       "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:     "test-offering",
				ImageUrl: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: ImageUrl"))
		})

		It("should fail when DocumentationUrl is not a URL", func() {
			definition := broker.ServiceDefinition{
				Id:               "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:             "test-offering",
				DocumentationUrl: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: DocumentationUrl"))
		})

		It("should fail when SupportUrl is not a URL", func() {
			definition := broker.ServiceDefinition{
				Id:         "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name:       "test-offering",
				SupportUrl: "some-non-url",
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("field must be a URL: SupportUrl"))
		})

		It("should fail when ProvisionInputVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				ProvisionInputVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): ProvisionInputVariables[0].details, ProvisionInputVariables[0].field_name"))
		})

		It("should fail when ProvisionComputedVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				ProvisionComputedVariables: []varcontext.DefaultVariable{
					{
						Overwrite: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): ProvisionComputedVariables[0].default, ProvisionComputedVariables[0].name"))
		})

		It("should fail when BindInputVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				BindInputVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): BindInputVariables[0].details, BindInputVariables[0].field_name"))
		})

		It("should fail when PlanVariables is not a valid", func() {
			definition := broker.ServiceDefinition{
				Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
				Name: "test-offering",
				PlanVariables: []broker.BrokerVariable{
					{
						ProhibitUpdate: true,
					},
				},
			}

			err := definition.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("missing field(s): PlanVariables[0].details, PlanVariables[0].field_name"))
		})

		Context("validate plans", func() {
			It("should fail when plan is missing name", func() {
				definition := broker.ServiceDefinition{
					Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
					Name: "test-offering",
					Plans: []broker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID: "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
							},
						},
					},
				}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("missing field(s): Plans[0].Name"))
			})

			It("should fail when plan is missing id", func() {
				definition := broker.ServiceDefinition{
					Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
					Name: "test-offering",
					Plans: []broker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								Name: "test-plan",
							},
						},
					},
				}

				err := definition.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("field must be a UUID: Plans[0].Id"))
			})

			Context("plan duplication", func() {
				It("should fail when plan id is duplicated across the offering", func() {
					definition := broker.ServiceDefinition{
						Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
						Name: "test-offering",
						Plans: []broker.ServicePlan{
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan-1",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan-2",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
						},
					}

					err := definition.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicated value, must be unique: c4edb10c-0434-11ec-8cbb-a777f76cf2ac: Plans[1].Id"))
				})

				It("should fail when plan name is duplicated across the offering", func() {
					definition := broker.ServiceDefinition{
						Id:   "55ad8194-0431-11ec-948a-63ff62e94b14",
						Name: "test-offering",
						Plans: []broker.ServicePlan{
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan",
									ID:   "c4edb10c-0434-11ec-8cbb-a777f76cf2ac",
								},
							},
							{
								ServicePlan: domain.ServicePlan{
									Name: "test-plan",
									ID:   "f71ef2c0-0435-11ec-a1e1-17b59e2ff283",
								},
							},
						},
					}

					err := definition.Validate()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("duplicated value, must be unique: test-plan: Plans[1].Name"))
				})
			})
		})

	})

	Describe("AllowedUpdate", func() {
		serviceDefinition := broker.ServiceDefinition{
			ProvisionInputVariables: []broker.BrokerVariable{
				{
					FieldName:      "prohibited",
					ProhibitUpdate: true,
				},
				{
					FieldName: "allowed",
				},
			},
		}

		DescribeTable("returns the correct result",
			func(params map[string]interface{}, expected bool) {
				actual, err := serviceDefinition.AllowedUpdate(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(Equal(expected))
			},
			Entry("allowed", map[string]interface{}{"allowed": "some_val"}, true),
			Entry("prohibited", map[string]interface{}{"prohibited": "some_val"}, false),
			Entry("empty", nil, true),
		)
	})
})
