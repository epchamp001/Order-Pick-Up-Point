package config

type AllowedConfig struct {
	Cities       map[string]bool `mapstructure:"cities"`
	ProductTypes map[string]bool `mapstructure:"product_types"`
	Roles        map[string]bool `mapstructure:"roles"`
}
