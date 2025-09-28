package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Kafka  KafkaConfig  `yaml:"kafka"`
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	LogLevel string `yaml:"log_level"`
}

type KafkaConfig struct {
	Host                 string `yaml:"host"`
	Port                 int    `yaml:"port"`
	OrderTopic           string `yaml:"order_topic"`
	OrderConsumerGroupID string `yaml:"consumer_group_id"`
	Brokers              string `yaml:"brokers"`
}

func LoadConfig(filename string) (*Config, error) {
	cleanPath := filepath.Clean(filename)

	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := &Config{}
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
