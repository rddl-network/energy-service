package config

import (
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `toml:"server"`
	InfluxDB   InfluxDBConfig   `toml:"influxdb"`
	Planetmint PlanetmintConfig `toml:"planetmint"`
}

type PlanetmintConfig struct {
	Actor   string `toml:"actor"`
	ChainID string `toml:"chain-id"`
	RPCHost string `toml:"rpc-host"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port     int    `toml:"port"`      // Port for the HTTP server
	LogLevel string `toml:"log-level"` // Log level: debug, info, warn, error
	DataFile string `toml:"data-file"` // Path to the data file
}

// InfluxDBConfig holds InfluxDB-related configuration
type InfluxDBConfig struct {
	URL    string `toml:"url"`    // InfluxDB URL
	Token  string `toml:"token"`  // InfluxDB authentication token
	Org    string `toml:"org"`    // InfluxDB organization
	Bucket string `toml:"bucket"` // InfluxDB bucket
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     8080,
			LogLevel: "info",
			DataFile: "energy_data.json",
		},
		InfluxDB: InfluxDBConfig{
			URL:    "http://localhost:8086",
			Token:  "",
			Org:    "",
			Bucket: "",
		},
		Planetmint: PlanetmintConfig{
			Actor:   "plmnt17keyseuam6qz4t49lg0e75a8y7jvcj03fn635z",
			ChainID: "testnetwork",
			RPCHost: "localhost:9090",
		},
	}
}

var config *Config

// LoadConfig loads the configuration from a TOML file
func LoadConfig(filePath string) (*Config, error) {
	if config != nil {
		return config, nil
	}

	cfg := DefaultConfig()

	data, err := os.ReadFile(filePath)
	if err == nil {
		err = toml.Unmarshal(data, cfg) // Unmarshal over defaults
		if err != nil {
			log.Printf("Error loading config file %s: %v", filePath, err)
			return nil, err
		}
	}

	config = cfg
	return config, nil
}

// GetConfig provides access to the loaded configuration
func GetConfig() *Config {
	if config == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig first.")
	}
	return config
}
