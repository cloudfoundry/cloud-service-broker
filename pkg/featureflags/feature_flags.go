package featureflags

import "github.com/spf13/viper"

const (
	TfUpgradeEnabled = "brokerpak.terraform.upgrades.enabled"
)

func init() {
	viper.BindEnv(TfUpgradeEnabled, "TERRAFORM_UPGRADES_ENABLED")
	viper.SetDefault(TfUpgradeEnabled, false)
}
