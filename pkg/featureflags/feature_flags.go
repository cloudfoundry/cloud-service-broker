package featureflags

import "github.com/spf13/viper"

type FeatureFlagName string

const (
	TfUpgradeEnabled                 FeatureFlagName = "brokerpak.terraform.upgrades.enabled"
	DynamicHCLEnabled                FeatureFlagName = "brokerpak.updates.enabled"
	DisableRequestPropertyValidation FeatureFlagName = "request.property.validation.disabled"
)

func init() {
	for ffName, varName := range map[FeatureFlagName]string{
		TfUpgradeEnabled:                 "TERRAFORM_UPGRADES_ENABLED",
		DynamicHCLEnabled:                "BROKERPAK_UPDATES_ENABLED",
		DisableRequestPropertyValidation: "CSB_DISABLE_REQUEST_PROPERTY_VALIDATION",
	} {
		viper.BindEnv(string(ffName), varName)
		viper.SetDefault(string(ffName), false)
	}
}

func Enabled(name FeatureFlagName) bool {
	return viper.GetBool(string(name))
}
