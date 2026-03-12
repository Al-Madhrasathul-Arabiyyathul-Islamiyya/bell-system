// config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Storage   StorageConfig
	CORS      CORSConfig
	WebSocket WebSocketConfig
}

type ServerConfig struct {
	Port         int
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn int // in minutes
}

type CORSConfig struct {
	AllowedOrigins []string
	MaxAge         int // seconds
}

type StorageConfig struct {
	AudioDir string
}

type WebSocketConfig struct {
	PingInterval   int // seconds, default 30
	PongTimeout    int // seconds, default 10
	MaxMessageSize int // bytes, default 512
}

// LoadConfig loads the configuration from config.toml
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
