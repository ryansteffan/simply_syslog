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

type SocketServerConfig struct {
	Enabled        bool   `json:"enabled"`
	Port           string `json:"port"`
	BindAddress    string `json:"bind_address"`
	MaxMessageSize int    `json:"max_message_size,omitempty"`
}

// Config defines the configuration for the syslog server.
type Config struct {
	Version           string             `json:"version"`
	UDPServer         SocketServerConfig `json:"udp_server"`
	TCPServer         SocketServerConfig `json:"tcp_server"`
	TLSServer         SocketServerConfig `json:"tls_server"`
	SelfLoggingLevel  applogger.LogLevel `json:"self_logging_level"`
	BufferMaxItems    int                `json:"buffer_max_items"`
	BufferMaxLifetime int                `json:"buffer_max_lifetime"`
}

// The default config that should be generated.
var defaultConfig = &Config{
	Version: "1.0.0",
	UDPServer: SocketServerConfig{
		Enabled:        true,
		Port:           "514",
		BindAddress:    "0.0.0.0",
		MaxMessageSize: 1024,
	},
	TCPServer: SocketServerConfig{
		Enabled:        false,
		Port:           "514",
		BindAddress:    "0.0.0.0",
		MaxMessageSize: 1024,
	},
	TLSServer: SocketServerConfig{
		Enabled:        false,
		Port:           "6514",
		BindAddress:    "0.0.0.0",
		MaxMessageSize: 1024,
	},
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
				Version: "1.0.0",

				UDPServer: SocketServerConfig{
					Enabled: func() bool {
						udpEnabledEnv := os.Getenv("UDP_SERVER_ENABLED")
						udpEnabled, err := strconv.ParseBool(udpEnabledEnv)
						if err != nil {
							return defaultConfig.UDPServer.Enabled
						}
						return udpEnabled
					}(),
					Port:        os.Getenv("UDP_PORT"),
					BindAddress: os.Getenv("UDP_BIND_ADDRESS"),
					MaxMessageSize: func() int {
						sizeEnv := os.Getenv("UDP_MAX_MESSAGE_SIZE")
						sizeInt, err := strconv.Atoi(sizeEnv)
						if err != nil {
							return defaultConfig.UDPServer.MaxMessageSize
						}
						return sizeInt
					}(),
				},

				TCPServer: SocketServerConfig{
					Enabled: func() bool {
						tcpEnabledEnv := os.Getenv("TCP_SERVER_ENABLED")
						tcpEnabled, err := strconv.ParseBool(tcpEnabledEnv)
						if err != nil {
							return defaultConfig.TCPServer.Enabled
						}
						return tcpEnabled
					}(),
					Port:        os.Getenv("TCP_PORT"),
					BindAddress: os.Getenv("TCP_BIND_ADDRESS"),
					MaxMessageSize: func() int {
						sizeEnv := os.Getenv("TCP_MAX_MESSAGE_SIZE")
						sizeInt, err := strconv.Atoi(sizeEnv)
						if err != nil {
							return defaultConfig.TCPServer.MaxMessageSize
						}
						return sizeInt
					}(),
				},

				TLSServer: SocketServerConfig{
					Enabled: func() bool {
						tlsEnabledEnv := os.Getenv("TLS_SERVER_ENABLED")
						tlsEnabled, err := strconv.ParseBool(tlsEnabledEnv)
						if err != nil {
							return defaultConfig.TLSServer.Enabled
						}
						return tlsEnabled
					}(),
					Port:        os.Getenv("TLS_PORT"),
					BindAddress: os.Getenv("TLS_BIND_ADDRESS"),
					MaxMessageSize: func() int {
						sizeEnv := os.Getenv("TLS_MAX_MESSAGE_SIZE")
						sizeInt, err := strconv.Atoi(sizeEnv)
						if err != nil {
							return defaultConfig.TLSServer.MaxMessageSize
						}
						return sizeInt
					}(),
				},

				SelfLoggingLevel: applogger.LogLevel(logLevelInt),

				BufferMaxItems: func() int {
					itemsEnv := os.Getenv("BUFFER_MAX_ITEMS")
					itemsInt, err := strconv.Atoi(itemsEnv)
					if err != nil {
						return defaultConfig.BufferMaxItems
					}
					return itemsInt
				}(),

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
