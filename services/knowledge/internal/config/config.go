// Package config is responsible for loading app's configuration from .env files
package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds configs for logger, gRPC server, connection to Neo4j database
type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL"`

	GRPC struct {
		Host string `envconfig:"HOST"`
		Port int    `envconfig:"PORT"`
	} `envconfig:"GRPC"`

	Neo4j struct {
		URI      string `envconfig:"URI"`
		Username string `envconfig:"USERNAME"`
		Password string `envconfig:"PASSWORD"`
		Database string `envconfig:"DATABASE"`
	} `envconfig:"NEO4J"`
}

// MustLoad loads config, if fails - panics
func MustLoad() Config {
	err := godotenv.Load()
	if err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: no .env file found, relying on system environment variables")
		}
	}

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
