package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Agent struct {
		Name    string `mapstructure:"name"`
		Country string `mapstructure:"country"`
	} `mapstructure:"agent"`

	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
		Topics  struct {
			Tasks   string `mapstructure:"tasks"`
			Results string `mapstructure:"results"`
			Logs    string `mapstructure:"logs"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`

	Server struct {
		HealthPort string `mapstructure:"health_port"`
	} `mapstructure:"server"`

	Backend struct {
		URL               string `mapstructure:"url"`
		RegistrationToken string `mapstructure:"registration_token"`
	} `mapstructure:"backend"`

	Checks struct {
		HTTPTimeout int `mapstructure:"http_timeout"`
		PingTimeout int `mapstructure:"ping_timeout"`
		TCPTimeout  int `mapstructure:"tcp_timeout"`
		DNSTimeout  int `mapstructure:"dns_timeout"`
	} `mapstructure:"checks"`
}

func Load() (*Config, error) {
	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetHTTPTimeout() time.Duration {
	return time.Duration(c.Checks.HTTPTimeout) * time.Second
}

func (c *Config) GetPingTimeout() time.Duration {
	return time.Duration(c.Checks.PingTimeout) * time.Second
}

func (c *Config) GetTCPTimeout() time.Duration {
	return time.Duration(c.Checks.TCPTimeout) * time.Second
}

func (c *Config) GetDNSTimeout() time.Duration {
	return time.Duration(c.Checks.DNSTimeout) * time.Second
}
