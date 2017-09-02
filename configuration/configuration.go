package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Config is a inner representation of the application's options
type Config struct {
	ZKillBoardURL    string
	ZKillBoardAPIURL string
	SlackChannel     string
	CorporationID    string
	AllianceID       string
	SlackUserName    string
	SlackIcon        string
	WebHookURL       string
}

// isMissing validates Configuration
func (c *Config) isMissing() bool {
	switch {
	case len(c.CorporationID) == 0 && len(c.AllianceID) == 0:
		return true
	case len(c.ZKillBoardURL) == 0 || len(c.ZKillBoardAPIURL) == 0:
		return true
	case len(c.SlackChannel) == 0 || len(c.WebHookURL) == 0:
		return true
	case len(c.SlackUserName) == 0 || len(c.SlackIcon) == 0:
		return true
	default:
		return false
	}
}

//NewConfig creates new config
func NewConfig() *Config {
	return &Config{}
}

// LoadConfiguration reading and loading configuration to Config variable
func (c *Config) LoadConfiguration(configFilePath string) {
	var err error
	err = c.readConfigFromJSON(configFilePath)
	if err != nil {
		fmt.Println("ERROR: ", err)
		log.Fatal(`Remote tools not found. Please specify configuration`)
		os.Exit(1)
	}
}

//readConfigFromJSON read configuration from conf.json
func (c *Config) readConfigFromJSON(configFilePath string) error {
	log.Printf("Looking for JSON config file in: %s\n", configFilePath)

	contents, err := ioutil.ReadFile(configFilePath)
	if err == nil {
		reader := bytes.NewBuffer(contents)
		err = json.NewDecoder(reader).Decode(&c)
	}
	if err != nil {
		log.Printf("Reading configuration from JSON (%s) failed: %s\n", configFilePath, err)
	}
	if c.isMissing() {
		err = errors.New("Configuration is missing")
	}
	return err
}
