package config

import (
	"encoding/json"
	"os"
	"sync"
)

var writerConfigPath = "./config/writer_config.json"

var writerConfig *WriterConfig
var writerConfigMutex sync.Mutex

type WriterConfig struct {
	Version string   `json:"version"`
	Writers []Writer `json:"writers"`
}

type Writer struct {
	Name    string            `json:"name"`
	Enabled bool              `json:"enabled"`
	Options map[string]string `json:"options"`
}

var defaultWriterConfig = &WriterConfig{
	Version: "1.0.0",
	Writers: []Writer{
		{
			Name:    "console",
			Enabled: true,
			Options: map[string]string{
				"format": "raw",
			},
		},
		{
			Name:    "file",
			Enabled: false,
			Options: map[string]string{
				"path": "./logs/app.log",
			},
		},
		{
			Name:    "sqlite",
			Enabled: false,
			Options: map[string]string{
				"path": "/syslog/syslog.db",
			},
		},
	},
}

func SaveWriterConfig(config *WriterConfig) error {
	writerConfigMutex.Lock()
	defer writerConfigMutex.Unlock()
	file, err := os.Create(writerConfigPath)
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

	writerConfig = config
	return nil
}

func loadWriterConfig() (*WriterConfig, error) {
	if _, err := os.Stat(writerConfigPath); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(writerConfigPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config WriterConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetWriterConfig() (*WriterConfig, error) {
	writerConfigMutex.Lock()
	defer writerConfigMutex.Unlock()

	if writerConfig != nil {
		return writerConfig, nil
	}

	var err error
	writerConfig, err = loadWriterConfig()
	if err != nil {
		return nil, err
	}
	return writerConfig, nil
}

func GenerateWriterConfig(fromENV bool) error {
	if _, err := os.Stat(writerConfigPath); os.IsNotExist(err) {
		if fromENV {
			// Make the initial config from the environment variables
		} else {
			// Save the default config to the file system
			return SaveWriterConfig(defaultWriterConfig)
		}
	}
	return nil
}
