package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

// globalConfig is the private application level config,
// access must be done via GetConfig and SaveConfig.
var globalConfig *Config

// globalConfigMutex is used to synchronize access to the globalConfig variable.
var globalConfigMutex sync.Mutex

// configPath is the path to the configuration file on the file system.
var configPath = "./config/config.json"

// var regexConfigPath = "./config/regex_config.json"
// var writerConfigPath = "./config/writer_config.json"

// ServerMode defines the mode in which the server can run.
type ServerMode string

const (
	// ServerModeUDP runs the server in UDP mode, listening for syslog messages on the specified UDP port.
	ServerModeUDP ServerMode = "udp"
	// ServerModeTCP runs the server in TCP mode, listening for syslog messages on the specified TCP port.
	ServerModeTCP ServerMode = "tcp"
	// ServerModeTLS runs the server in TLS mode, listening for syslog messages on the specified TLS port.
	ServerModeTLS ServerMode = "tls"
	// ServerModeAll runs the server in all modes, listening for syslog messages on the specified
	// UDP, TCP, and TLS ports.
	ServerModeAll ServerMode = "all"
)

// Config defines the configuration for the syslog server.
type Config struct {
	Version           string             `json:"version"`
	ServerMode        ServerMode         `json:"server_mode"`
	BindAddress       string             `json:"bind_address"`
	UDPPort           string             `json:"udp_port"`
	TCPPort           string             `json:"tcp_port"`
	TLSPort           string             `json:"tls_port"`
	SelfLoggingLevel  applogger.LogLevel `json:"self_logging_level"`
	BufferMaxItems    int                `json:"buffer_max_items"`
	BufferMaxLifetime int                `json:"buffer_max_lifetime"`
}

// The default config that should be generated.
var defaultConfig = &Config{
	Version:           "1.0.0",
	ServerMode:        ServerModeUDP,
	BindAddress:       "0.0.0.0",
	UDPPort:           "514",
	TCPPort:           "514",
	TLSPort:           "6514",
	SelfLoggingLevel:  applogger.DEBUG,
	BufferMaxItems:    1024,
	BufferMaxLifetime: 15,
}

// SaveConfig saves the configuration to the file system and updates the globalConfig variable
func SaveConfig(config *Config) error {
	globalConfigMutex.Lock()
	defer globalConfigMutex.Unlock()
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	json, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(json)
	if err != nil {
		return err
	}

	globalConfig = config

	return nil
}

// loadConfig loads the configuration from the file system
func loadConfig() (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config Config

	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetConfig returns the global configuration, loading it from the file system if it hasn't been loaded yet
func GetConfig() (*Config, error) {
	globalConfigMutex.Lock()
	defer globalConfigMutex.Unlock()

	if globalConfig != nil {
		return globalConfig, nil
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	globalConfig = cfg
	return globalConfig, nil
}

// GenerateConfig generates the configuration file if it doesn't exist, either from the environment
// variables or from the default config
//
// It is intended for this to be called at the start of the application but it can be called later to no effect.
func GenerateConfig(fromENV bool) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if fromENV {
			// Make the initial config from the environment variables
			logLevelEnv := os.Getenv("SELF_LOGGING_LEVEL")
			logLevelInt, err := strconv.Atoi(logLevelEnv)
			if err != nil {
				return errors.New("invalid SELF_LOGGING_LEVEL environment variable")
			}
			globalConfig = &Config{
				Version:          "1.0.0",
				ServerMode:       ServerMode(os.Getenv("SERVER_MODE")),
				BindAddress:      os.Getenv("BIND_ADDRESS"),
				UDPPort:          os.Getenv("UDP_PORT"),
				TCPPort:          os.Getenv("TCP_PORT"),
				TLSPort:          os.Getenv("TLS_PORT"),
				SelfLoggingLevel: applogger.LogLevel(logLevelInt),

				// Parse out the max items
				BufferMaxItems: func() int {
					itemsEnv := os.Getenv("BUFFER_MAX_ITEMS")
					itemsInt, err := strconv.Atoi(itemsEnv)
					if err != nil {
						return defaultConfig.BufferMaxItems
					}
					return itemsInt
				}(),
				// Parse out the max lifetime
				BufferMaxLifetime: func() int {
					lifetimeEnv := os.Getenv("BUFFER_MAX_LIFETIME")
					lifetimeInt, err := strconv.Atoi(lifetimeEnv)
					if err != nil {
						return defaultConfig.BufferMaxLifetime
					}
					return lifetimeInt
				}(),
			}
			return SaveConfig(globalConfig)
		} else {
			return SaveConfig(defaultConfig)
		}
	}
	return nil
}
