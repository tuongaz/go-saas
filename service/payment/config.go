package payment

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	StripePrivateKey string `mapstructure:"GOS_PAYMENT_STRIPE_PRIVATE_KEY"`
}

func SetDefault(key string, value any) {
	viper.SetDefault(key, value)
}

func newConfig() (*Config, error) {
	viper.AutomaticEnv()

	SetDefault("GOS_PAYMENT_STRIPE_PRIVATE_KEY", "")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return &cfg, nil
}
