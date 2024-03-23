package config

type Config struct {
	Host     string `json:"host" envconfig:"HOST"`
	HostPort string `json:"hostPort" envconfig:"HOST_PORT"`
}

func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	if err := loadEnvConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
