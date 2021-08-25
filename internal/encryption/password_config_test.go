package encryption_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption Config", func() {
	Describe("PasswordConfigs", func() {
		Describe("Validate", func() {
			It("should error when no encryption keys are provided", func() {
				passwordConfigs := encryption.PasswordConfigs{}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("no encryption keys were provided"))
			})

			It("should error when one of the password configs is invalid", func() {
				passwordConfigs := encryption.PasswordConfigs{
					encryption.PasswordConfig{
						ID: "test-id",
					},
				}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("field must be a UUID: Key[0].guid\nmissing field(s): Key[0].encryption_key.secret, Key[0].label"))
			})

			It("should error when same id is used by several password configs", func() {
				passwordConfigs := encryption.PasswordConfigs{
					encryption.PasswordConfig{
						ID:      "40062468-05bc-11ec-822c-73dd987c0dd6",
						Label:   "some-test-label",
						Primary: true,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
					encryption.PasswordConfig{
						ID:      "40062468-05bc-11ec-822c-73dd987c0dd6",
						Label:   "some-other-test-label",
						Primary: false,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
				}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("duplicated value, must be unique: 40062468-05bc-11ec-822c-73dd987c0dd6: Key[1].guid"))
			})

			It("should error when same label is used by several password configs", func() {
				passwordConfigs := encryption.PasswordConfigs{
					encryption.PasswordConfig{
						ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
						Label:   "same-test-label",
						Primary: true,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
					encryption.PasswordConfig{
						ID:      "40062468-05bc-11ec-822c-73dd987c0dd6",
						Label:   "same-test-label",
						Primary: false,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
				}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("duplicated value, must be unique: same-test-label: Key[1].label"))
			})

			It("should error when no password is marked as primary", func() {
				passwordConfigs := encryption.PasswordConfigs{
					encryption.PasswordConfig{
						ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
						Label:   "some-test-label",
						Primary: false,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
					encryption.PasswordConfig{
						ID:      "40062468-05bc-11ec-822c-73dd987c0dd6",
						Label:   "some-other-test-label",
						Primary: false,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
				}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("no encryption key is marked as primary"))
			})

			It("should error when several passwords are marked as primary", func() {
				passwordConfigs := encryption.PasswordConfigs{
					encryption.PasswordConfig{
						ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
						Label:   "some-test-label",
						Primary: true,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
					encryption.PasswordConfig{
						ID:      "40062468-05bc-11ec-822c-73dd987c0dd6",
						Label:   "some-other-test-label",
						Primary: true,
						Password: encryption.Password{
							Secret: "some-super-secret-password",
						},
					},
				}

				err := passwordConfigs.Validate()
				Expect(err).To(MatchError("several encryption keys are marked as primary"))
			})
		})
	})

	Describe("PasswordConfig", func() {
		Describe("Validate", func() {
			It("should error when guid is missing", func() {
				password := encryption.PasswordConfig{
					ID:      "test",
					Label:   "some-label",
					Primary: false,
					Password: encryption.Password{
						Secret: "something-extremely-secret",
					},
				}

				err := password.Validate()
				Expect(err.Error()).To(Equal("field must be a UUID: guid"))
			})

			It("should error when label is missing", func() {
				password := encryption.PasswordConfig{
					ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
					Primary: false,
					Password: encryption.Password{
						Secret: "something-extremely-secret",
					},
				}

				err := password.Validate()
				Expect(err.Error()).To(Equal("missing field(s): label"))
			})

			It("should error when label is invalid size", func() {
				password := encryption.PasswordConfig{
					ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
					Label:   "test",
					Primary: false,
					Password: encryption.Password{
						Secret: "something-extremely-secret",
					},
				}

				err := password.Validate()
				Expect(err.Error()).To(Equal("field must be 5-1024 chars long: test: label"))
			})

			It("should error when secret is missing", func() {
				password := encryption.PasswordConfig{
					ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
					Label:   "test-label",
					Primary: false,
					Password: encryption.Password{
						Secret: "",
					},
				}

				err := password.Validate()
				Expect(err.Error()).To(Equal("missing field(s): encryption_key.secret"))
			})

			It("should error when secret is invalid size", func() {
				password := encryption.PasswordConfig{
					ID:      "f4a34ccc-05ba-11ec-bcb7-fb8b57c059aa",
					Label:   "test-label",
					Primary: false,
					Password: encryption.Password{
						Secret: "short-secret",
					},
				}

				err := password.Validate()
				Expect(err.Error()).To(Equal("field must be 20-1024 chars long: short-secret: encryption_key.secret"))
			})
		})
	})
})
