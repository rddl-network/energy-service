package config

import (
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `toml:"server"`
	InfluxDB InfluxDBConfig `toml:"influxdb"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port     int    `toml:"port"`      // Port for the HTTP server
	LogLevel string `toml:"log_level"` // Log level: debug, info, warn, error
	DataFile string `toml:"data_file"` // Path to the data file
}

// InfluxDBConfig holds InfluxDB-related configuration
type InfluxDBConfig struct {
	URL    string `toml:"url"`    // InfluxDB URL
	Token  string `toml:"token"`  // InfluxDB authentication token
	Org    string `toml:"org"`    // InfluxDB organization
	Bucket string `toml:"bucket"` // InfluxDB bucket
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     8080,
			LogLevel: "info",
			DataFile: "energy_data.json",
		},
		InfluxDB: InfluxDBConfig{
			URL:    "https://eu-central-1-1.aws.cloud2.influxdata.com",
			Token:  "TBUV2ciRQYeM2KUOGJmt1V0c7jv0CqxYhcSaGpELe3YLnc3Tc2dcQEAbrZplmDcb-HSBLbPr9kAXPHpvPf8ezw==",
			Org:    "713a74226aae814d",
			Bucket: "0ffa8c3c2d0957a8",
		},
	}
}

var config *Config

// LoadConfig loads the configuration from a TOML file
func LoadConfig(filePath string) (*Config, error) {
	if config != nil {
		return config, nil
	}

	cfg := defaultConfig()

	data, err := os.ReadFile(filePath)
	if err == nil {
		_ = toml.Unmarshal(data, cfg) // Unmarshal over defaults
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
