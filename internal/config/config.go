package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds the application's configuration
type Config struct {
	Prometheus        PrometheusConfig `yaml:"prometheus"`
	Slack             SlackConfig      `yaml:"slack"`
	Thresholds        ThresholdsConfig `yaml:"thresholds"`
	DisableCollection bool             `yaml:"disable_collection"`
}

// PrometheusConfig holds Prometheus-related configuration
type PrometheusConfig struct {
	ListenAddress string `yaml:"listen_address"`
}

// SlackConfig holds Slack-related configuration
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

// ThresholdsConfig holds alerting thresholds
type ThresholdsConfig struct {
	CPUPercent float64 `yaml:"cpu_percent"`
	MemPercent float64 `yaml:"mem_percent"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Prometheus.ListenAddress == "" {
		return fmt.Errorf("prometheus listen_address is required")
	}
	if c.Slack.WebhookURL == "" {
		return fmt.Errorf("slack webhook_url is required")
	}
	if c.Thresholds.CPUPercent <= 0 || c.Thresholds.CPUPercent > 100 {
		return fmt.Errorf("cpu_percent must be between 0 and 100")
	}
	if c.Thresholds.MemPercent <= 0 || c.Thresholds.MemPercent > 100 {
		return fmt.Errorf("mem_percent must be between 0 and 100")
	}
	return nil
}

// LoadConfig reads the configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %w", configPath, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
