// This package contains the required code to load the
// JSON server config into the application.
package config

import (
	"encoding/json"
	"errors"
	"os"
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
	Max_Tcp_Connections string `json:"max_tcp_connections"`
	Buffer_Length       string `json:"buffer_length"`
	Buffer_Lifespan     string `json:"buffer_lifespan"`
	Max_Message_Size    string `json:"max_message_size"`
	Syslog_Path         string `json:"syslog_path"`
	Debug_Messages      string `json:"debug_messages"`
}

/*
Loads the data from a JSON config file into a Config struct,
and then returns a pointer to the new config.

Returns an error if the file does not exist,
or there was an error with the unmarshaling of
the file.
*/
func LoadConfig(path string) (*Config, error) {

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

func (conf *Config) SaveConfig() error {
	data, err := json.Marshal(conf.Data)
	if err != nil {
		return errors.New("there was an error encoding the save file")
	}

	os.WriteFile(conf.FileLocation, data, 0644)
	return nil
}
