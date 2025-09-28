package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host                 string `yaml:"host"`
		Port                 string `yaml:"port"`
		CartCap              int    `yaml:"cart_cap"`
		RetryTimeout         int    `yaml:"retry_timeout"`
		Timeout              int    `yaml:"timeout"`
		Workers              int    `yaml:"workers"`
		LogLevel             string `yaml:"log_level"`
		CheckStorageInterval int    `yaml:"check_storage_interval"`
	} `yaml:"service"`
	ProductService struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		Token   string `yaml:"token"`
		Timeout int    `yaml:"timeout"`
		Limit   int    `yaml:"limit"`
		Burst   int    `yaml:"burst"`
	} `yaml:"product_service"`
	LomsService struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"loms_service"`
	Jaeger struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"jaeger"`
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
