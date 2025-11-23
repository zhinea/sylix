package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Log struct {
		Level      string `yaml:"level"`
		Filename   string `yaml:"filename"`
		MaxSize    int    `yaml:"max_size"` // MB
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"` // days
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

func LoadAgentConfig(path string) (*AgentConfig, error) {
	// Default configuration
	config := &AgentConfig{}
	config.Server.Port = 8083
	config.Server.Host = "0.0.0.0"
	config.Log.Level = "info"
	config.Log.Filename = "sylix-agent.log"
	config.Log.MaxSize = 10
	config.Log.MaxBackups = 3
	config.Log.MaxAge = 28
	config.Log.Compress = true

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return defaults if file doesn't exist
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
