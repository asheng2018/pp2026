package config

import (
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/merchant/model"
)

// DefaultFeeConfigs returns preset fee configurations.
func DefaultFeeConfigs() map[string]*model.FeeConfig {
	return map[string]*model.FeeConfig{
		"standard": {
			PayPalFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.30),
				RateFee:  decimal.NewFromFloat(0.044), // 4.4%
				MinFee:   decimal.NewFromFloat(0.50),
			},
			StripeFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.30),
				RateFee:  decimal.NewFromFloat(0.039), // 3.9%
				MinFee:   decimal.NewFromFloat(0.50),
			},
		},
		"premium": {
			PayPalFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.30),
				RateFee:  decimal.NewFromFloat(0.035), // 3.5%
				MinFee:   decimal.NewFromFloat(0.30),
			},
			StripeFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.30),
				RateFee:  decimal.NewFromFloat(0.029), // 2.9%
				MinFee:   decimal.NewFromFloat(0.30),
			},
		},
		"high_volume": {
			PayPalFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.30),
				RateFee:  decimal.NewFromFloat(0.029), // 2.9%
				MinFee:   decimal.NewFromFloat(0.25),
			},
			StripeFee: &model.GatewayFee{
				FixedFee: decimal.NewFromFloat(0.25),
				RateFee:  decimal.NewFromFloat(0.025), // 2.5%
				MinFee:   decimal.NewFromFloat(0.20),
			},
		},
	}
}

func GetDefaultFeeConfig() *model.FeeConfig {
	return DefaultFeeConfigs()["standard"]
}
