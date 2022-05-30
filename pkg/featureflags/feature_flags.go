package featureflags

import "github.com/spf13/viper"

const (
	TfUpgradeEnabled  = "brokerpak.terraform.upgrades.enabled"
	DynamicHCLEnabled = "brokerpak.updates.enabled"
)

func init() {
	viper.BindEnv(TfUpgradeEnabled, "TERRAFORM_UPGRADES_ENABLED")
	viper.SetDefault(TfUpgradeEnabled, false)

	viper.BindEnv(DynamicHCLEnabled, "BROKERPAK_UPDATES_ENABLED")
	viper.SetDefault(DynamicHCLEnabled, false)
}
