package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Service  Service      `yaml:"service"`
	MasterDB DBConfig     `yaml:"db_master"`
	ReplicDB DBConfig     `yaml:"db_replica"`
	Kafka    KafkaConfig  `yaml:"kafka"`
	Jaeger   JaegerConfig `yaml:"jaeger"`
}

type Service struct {
	Host           string `yaml:"host"`
	GRPCPort       int    `yaml:"grpc_port"`
	HTTPPort       int    `yaml:"http_port"`
	SwaggerPort    int    `yaml:"swagger_port"`
	Timeout        int    `yaml:"timeout"`
	HandlePeriod   int    `yaml:"handle_period"`
	LimitOutboxMsg int32  `yaml:"limit_outbox_msg"`
	LogLevel       string `yaml:"log_level"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"db_name"`
}

type KafkaConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	OrderTopic    string `yaml:"order_topic"`
	Brokers       string `yaml:"brokers"`
	RetryCountMsg int    `yaml:"retry_count_msg"`
}

type JaegerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
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
