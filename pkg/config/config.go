package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		LogURL        string        `yaml:"log_url"`
		Username      string        `yaml:"username"`
		Password      string        `yaml:"password"`
		CheckInterval time.Duration `yaml:"check_interval"`
		StateFile     string        `yaml:"state_file"`
	} `yaml:"server"`
	Sentry struct {
		DSN         string `yaml:"dsn"`
		Environment string `yaml:"environment"`
		Project     string `yaml:"project"`
	} `yaml:"sentry"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
