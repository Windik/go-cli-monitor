package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the configuration for the CLI monitor
type Config struct {
	CheckInterval int      `mapstructure:"check_interval"`
	Targets       []string `mapstructure:"targets"`
	DefaultTitle  string   `mapstructure:"default_title"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	viper.SetDefault("check_interval", 5)
	viper.SetDefault("targets", []string{"https://google.com"})
	viper.SetDefault("default_title", "CLI Monitor")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				return nil, fmt.Errorf("failed to write default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
