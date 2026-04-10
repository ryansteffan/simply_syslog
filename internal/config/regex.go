package config

import (
	"encoding/json"
	"os"
	"sync"
)

var regexConfig *RegexConfig
var regexConfigMutex sync.Mutex

var regexConfigPath = "./config/regex_config.json"

type RegexPattern struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type RegexConfig struct {
	Version string         `json:"version"`
	Regexes []RegexPattern `json:"regexes"`
}

var defaultRegexConfig = &RegexConfig{
	Version: "1.0.0",
	Regexes: []RegexPattern{
		{
			Name:    "RFC3164",
			Pattern: `<(?P<pri>\d{1,3})>(?P<timestamp>\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(?P<hostname>[\w\.\-]+)\s+(?P<message>.+)`,
		},
		{
			Name:    "RFC5424",
			Pattern: `<(?P<pri>\d{1,3})>(?P<version>\d{1,2})\s+(?P<timestamp>[\w\-\:\.]+)\s+(?P<hostname>[\w\.\-]+)\s+(?P<appname>[\w\.\-]+)\s+(?P<procid>[\w\.\-]+)\s+(?P<msgid>[\w\.\-]+)\s+(?P<message>.+)`,
		},
	},
}

func loadRegexConfig() (*RegexConfig, error) {
	if _, err := os.Stat(regexConfigPath); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(regexConfigPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config RegexConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveRegexConfig(config *RegexConfig) error {
	regexConfigMutex.Lock()
	defer regexConfigMutex.Unlock()
	file, err := os.Create(regexConfigPath)
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

	regexConfig = config
	return nil
}

func GetRegexConfig() (*RegexConfig, error) {
	regexConfigMutex.Lock()
	defer regexConfigMutex.Unlock()

	if regexConfig != nil {
		return regexConfig, nil
	}

	var err error
	regexConfig, err = loadRegexConfig()
	if err != nil {
		return nil, err
	}
	return regexConfig, nil
}

func GenerateRegexConfig(saveFromEnv bool) error {
	if _, err := os.Stat(regexConfigPath); os.IsNotExist(err) {
		if saveFromEnv {
			// Parse the environment variables to create the initial regex config
			regexVar := os.Getenv("REGEX_CONFIG")

			var regexes []RegexPattern
			err := json.Unmarshal([]byte(regexVar), &regexes)
			if err != nil {
				return err
			}

			regexConfig := &RegexConfig{
				Version: "1.0.0",
				Regexes: regexes,
			}
			return SaveRegexConfig(regexConfig)
		} else {
			return SaveRegexConfig(defaultRegexConfig)
		}
	}
	return nil
}
