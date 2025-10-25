package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env     string        `mapstructure:"env"`
	Agent   AgentConfig   `mapstructure:"agent"`
	Kafka   KafkaConfig   `mapstructure:"kafka"`
	Server  ServerConfig  `mapstructure:"server"`
	Backend BackendConfig `mapstructure:"backend"`
	Checks  ChecksConfig  `mapstructure:"checks"`
}

type AgentConfig struct {
	Name    string `mapstructure:"name"`
	Country string `mapstructure:"country"`
	Token   string `mapstructure:"token"`
}

type KafkaConfig struct {
	Brokers []string    `mapstructure:"brokers"`
	Topics  KafkaTopics `mapstructure:"topics"`
}

type KafkaTopics struct {
	Tasks   string `mapstructure:"tasks"`
	Results string `mapstructure:"results"`
	Logs    string `mapstructure:"logs"`
}

type ServerConfig struct {
	HealthPort string `mapstructure:"health_port"`
}

type BackendConfig struct {
	URL string `mapstructure:"url"`
}

type ChecksConfig struct {
	HTTPTimeout int `mapstructure:"http_timeout"`
	PingTimeout int `mapstructure:"ping_timeout"`
	TCPTimeout  int `mapstructure:"tcp_timeout"`
	DNSTimeout  int `mapstructure:"dns_timeout"`
}

func Load() (*Config, error) {

	setDefaults()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Agent defaults
	viper.SetDefault("env", "local")
	viper.SetDefault("agent.name", "monitoring-agent-01")
	viper.SetDefault("agent.country", "RU")
	viper.SetDefault("agent.token", "")

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topics.tasks", "agent-tasks")
	viper.SetDefault("kafka.topics.results", "check-results")
	viper.SetDefault("kafka.topics.logs", "agent-logs")

	// Server defaults
	viper.SetDefault("server.health_port", "8081")

	// Checks defaults
	viper.SetDefault("checks.http_timeout", 10)
	viper.SetDefault("checks.ping_timeout", 5)
	viper.SetDefault("checks.tcp_timeout", 5)
	viper.SetDefault("checks.dns_timeout", 5)
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
