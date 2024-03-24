package config

type Config struct {
	ENV           string `json:"env" envconfig:"ENV"`
	Host          string `json:"host" envconfig:"HOST"`
	HostPort      string `json:"hostPort" envconfig:"HOST_PORT"`
	RedisAddr     string `json:"redisAddr" envconfig:"REDIS_ADDR"`
	RedisUserName string `json:"redisUserName" envconfig:"REDIS_USERNAME"`
	RedisPassword string `json:"redisPassword" envconfig:"REDIS_PASSWORD"`
	ZeroXApiKey   string `json:"zeroXApiKey" envconfig:"ZEROX_KEY"`
	RpcURL        string `json:"rpcURL" envconfig:"RPC_URL"`
	ChainID       string `json:"chainID" envconfig:"CHAIN_ID"`
}

func NewConfigFromEnv() (*Config, error) {

	cfg := &Config{}
	if err := loadEnvConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
