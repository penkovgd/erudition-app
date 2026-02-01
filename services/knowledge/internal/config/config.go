// Package config is responsible for loading app's configuration from .env files
package config

import (
	"fmt"

	_ "github.com/joho/godotenv/autoload" // Autoload .env files
	"github.com/kelseyhightower/envconfig"
)

// Config holds configs for logger, gRPC server, connection to Neo4j database
type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL"`

	GRPC struct {
		Host string `envconfig:"HOST"`
		Port int    `envconfig:"PORT"`
	} `evnconfig:"GRPC"`

	Neo4j struct {
		URI      string `envconfig:"URI" required:"true"`
		Username string `envconfig:"USERNAME"`
		Password string `envconfig:"PASSWORD"`
		Database string `envconfig:"DATABASE"`
	} `envconfig:"NEO4J"`
}

// MustLoad loads config, if fails - panics
func MustLoad() Config {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		panic("failed to load config: " + err.Error())
	}

	return cfg
}

// GRPCAddress returns gRPC address in format: <host>:<port>
func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPC.Host, c.GRPC.Port)
}
