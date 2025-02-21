package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL" env-required:"true"`
	RedisURL    string `yaml:"redis_url" env:"REDIS_URL" env-required:"true"`
	Tls         Tls    `yaml:"tls"`
	GrpcPort    string `yaml:"grpc_port" env-required:"true"`
	HttpPort    string `yaml:"http_port" env-required:"true"`
	Auth        Auth   `yaml:"auth" env-required:"true"`
}

type Tls struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type Auth struct {
	JwtSecret       string `yaml:"jwt_secret" env-required:"true"`
	AccessTokenTTL  int    `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL int    `yaml:"refresh_token_ttl" env-required:"true"`
	TmaSecret       string `yaml:"tma_secret" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
