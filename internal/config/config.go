// This package contains the required code to load the
// JSON server config into the application.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/ryansteffan/simply_syslog/internal/utils"
)

type Config struct {
	FileLocation string
	Data         ConfigData
}

/*
Represents the JSON config file for the server.

Tags correspond to the field in the JSON file.
*/
type ConfigData struct {
	Protocol            string `json:"protocol"`
	Bind_Address        string `json:"bind_address"`
	Udp_Port            string `json:"udp_port"`
	Tcp_Port            string `json:"tcp_port"`
	Max_Tcp_Connections int    `json:"max_tcp_connections"`
	Buffer_Length       int    `json:"buffer_length"`
	Buffer_Lifespan     int    `json:"buffer_lifespan"`
	Max_Message_Size    int    `json:"max_message_size"`
	Syslog_Path         string `json:"syslog_path"`
	Debug_Messages      bool   `json:"debug_messages"`
}

/*
Loads in the config data for the server.

The config can be retrieved via two methods:

1) Using environment variables.

2) Using a json config file.

To load a json config file, pass the path to the config file.

When loading the config from environment variables, pass in the "ENV" string.

The path for the config will then be set to "ENV".
*/
func LoadConfig(path string) (*Config, error) {

	switch path {

	case "ENV":
		var envTypeError *error = new(error)

		conf := Config{
			FileLocation: "ENV",
			Data: ConfigData{
				Protocol:     utils.DefaultStringValue(os.Getenv("PROTOCOL"), "UDP"),
				Bind_Address: utils.DefaultStringValue(os.Getenv("BIND_ADDRESS"), "0.0.0.0"),
				Udp_Port:     utils.DefaultStringValue(os.Getenv("UDP_PORT"), "514"),
				Tcp_Port:     utils.DefaultStringValue(os.Getenv("TCP_PORT"), "514"),
				Syslog_Path:  utils.DefaultStringValue(os.Getenv("SYSLOG_PATH"), "/var/log/simply_syslog"),

				Max_Tcp_Connections: utils.InlineIntParse(
					utils.DefaultStringValue(os.Getenv("MAX_TCP_CONNECTIONS"), "10"),
					envTypeError,
				),

				Buffer_Length: utils.InlineIntParse(
					utils.DefaultStringValue(os.Getenv("BUFFER_LENGTH"), "32"),
					envTypeError,
				),

				Buffer_Lifespan: func() int {
					result, err := strconv.Atoi(utils.DefaultStringValue(os.Getenv("BUFFER_LIFESPAN"), "5"))
					if err != nil {
						*envTypeError = err
					}

					return result
				}(),
				Max_Message_Size: func() int {
					result, err := strconv.Atoi(utils.DefaultStringValue(os.Getenv("MAX_MESSAGE_SIZE"), "1024"))
					if err != nil {
						*envTypeError = err
					}

					return result
				}(),

				Debug_Messages: func() bool {
					result, err := strconv.ParseBool(utils.DefaultStringValue(os.Getenv("DEBUG_MESSAGES"), "True"))
					if err != nil {
						*envTypeError = err
					}

					return result
				}(),
			},
		}

		if envTypeError != nil {
			return nil, *envTypeError
		}

		return &conf, nil

	default:
		configFile, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.New("error reading in config file contents")
		}

		var config Config
		err = json.Unmarshal(configFile, &config.Data)
		if err != nil {
			return nil, errors.New("error parsing config file into struct")
		}

		config.FileLocation = path

		return &config, nil
	}
}

/*
Saves a config to it's path.

If the config is loaded from environment variables, then it will update variables.

If a json config was loaded, then the changes are written to the loaded file.
*/
func (conf *Config) SaveConfig() error {
	data, err := json.Marshal(conf.Data)
	if err != nil {
		return errors.New("there was an error encoding the save file")
	}

	os.WriteFile(conf.FileLocation, data, 0644)
	return nil
}
