package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	envFilePath string
}

type Option func(*envConfig)

func defaultEnvConfig() *envConfig {
	return &envConfig{
		envFilePath: ".env",
	}
}

func WithEnvPath(path string) Option {
	return func(cfg *envConfig) {
		cfg.envFilePath = path
	}
}

func loadEnvConfig(cfg any, options ...Option) error {
	defaultCfg := defaultEnvConfig()
	for _, option := range options {
		option(defaultCfg)
	}

	if err := godotenv.Overload(defaultCfg.envFilePath); err != nil {
		return fmt.Errorf("failed to read env file, %w", err)
	}

	if err := envconfig.Process("", cfg); err != nil {
		return fmt.Errorf("failed to fill config structure.en, %w", err)
	}

	return nil
}
