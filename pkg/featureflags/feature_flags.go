// Package featureflags is used to determine the state of feature flags
package featureflags

import "github.com/spf13/viper"

type FeatureFlagName string

const (
	TfUpgradeEnabled                 FeatureFlagName = "brokerpak.terraform.upgrades.enabled"
	DynamicHCLEnabled                FeatureFlagName = "brokerpak.updates.enabled"
	DisableRequestPropertyValidation FeatureFlagName = "request.property.validation.disabled"

	// DisableBindOutputMerging is a feature flag that enables the legacy behavior of merging the output from
	// the provision with the outputs from the binding. Instead, outputs required in the binding should be explicitly
	// copied rather than defaulting to being copied. This prevents potential security issues where a provision operation
	// may create and admin user, and the credentials for that admin user should not be leaked into the binding credentials.
	// This feature flag will be removed in the future, so it's better to update a brokerpak than to try and rely on this flag.
	DisableBindOutputMerging FeatureFlagName = "bind.output.merging.disabled"
)

func init() {
	for ffName, varName := range map[FeatureFlagName]string{
		TfUpgradeEnabled:                 "TERRAFORM_UPGRADES_ENABLED", // deprecated pattern - future variables should start CSB_
		DynamicHCLEnabled:                "BROKERPAK_UPDATES_ENABLED",  // deprecated pattern - future variables should start CSB_
		DisableRequestPropertyValidation: "CSB_DISABLE_REQUEST_PROPERTY_VALIDATION",
		DisableBindOutputMerging:         "CSB_DISABLE_BIND_OUTPUT_MERGING",
	} {
		viper.BindEnv(string(ffName), varName)
		viper.SetDefault(string(ffName), false)
	}
}

func Enabled(name FeatureFlagName) bool {
	return viper.GetBool(string(name))
}
