// Package featureflags is used to determine the state of feature flags
package featureflags

import "github.com/spf13/viper"

type FeatureFlagName string

const (
	TfUpgradeEnabled                 FeatureFlagName = "brokerpak.terraform.upgrades.enabled"
	DynamicHCLEnabled                FeatureFlagName = "brokerpak.updates.enabled"
	DisableRequestPropertyValidation FeatureFlagName = "request.property.validation.disabled"

	// EnableLegacyExamplesCommands enabled the old way of running example tests
	// IF YOU USE THIS, PLEASE RAISE AN ISSUE. Since the new way of running examples was added,
	// the authors don't expect anyone to use the legacy method. Please let us know if you need
	// the old way to stay around.
	EnableLegacyExamplesCommands    FeatureFlagName = "legacy.examples.enabled"
	DisableTfUpgradeProviderRenames FeatureFlagName = "brokerpak.terraform.upgrades.providerRenames.disabled"
)

var AllFeatureFlagEnvVars []string

func init() {
	for ffName, varName := range map[FeatureFlagName]string{
		TfUpgradeEnabled:                 "TERRAFORM_UPGRADES_ENABLED", // deprecated pattern - future variables should start CSB_
		DynamicHCLEnabled:                "BROKERPAK_UPDATES_ENABLED",  // deprecated pattern - future variables should start CSB_
		DisableRequestPropertyValidation: "CSB_DISABLE_REQUEST_PROPERTY_VALIDATION",
		EnableLegacyExamplesCommands:     "CSB_ENABLE_LEGACY_EXAMPLES_COMMANDS",
		DisableTfUpgradeProviderRenames:  "CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES",
	} {
		viper.BindEnv(string(ffName), varName)
		viper.SetDefault(string(ffName), false)
		AllFeatureFlagEnvVars = append(AllFeatureFlagEnvVars, varName)
	}
}

func Enabled(name FeatureFlagName) bool {
	return viper.GetBool(string(name))
}
