package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"log"
	"os"
)

// Config is a inner representation of the application's options
type Config struct {
	ZKillBoardUrl string
	SlackChannel  string
	SlackURL      string
	CorporationID string
	AllianceID    string
	SlackUserName string
	SlackIcon     string
}

// isMissing validates Configuration
func (c *Config) isMissing() bool {
	switch {
	case len(c.CorporationID) == 0 && len(c.AllianceID) == 0:
		return true
	case len(c.ZKillBoardUrl) == 0:
		return true
	case len(c.SlackChannel) == 0:
		return true
	case len(c.SlackURL) == 0:
		return true
	case len(c.SlackUserName) == 0:
		return true
	case len(c.SlackIcon) == 0:
		return true
	default:
		return false
	}
}

// Cfg is a package variable, which is populated during init() execution and shared to whole application
var Cfg Config

// LoadConfiguration reading and loading configuration to Config variable
func LoadConfiguration(configFilePath string) {
	var err error
	if len(configFilePath) != 0 {
		// read configuration from JSON
		err = readConfigFromJSON(configFilePath)
	} else {
		// read configuration from ENVIRONMENT
		err = readConfigFromENV()
	}
	if err != nil {
		log.Fatal(`Remote tools not found. Please specify configuration`)
		os.Exit(1)
	}
}

//readConfigFromJSON read configuration from conf.json
func readConfigFromJSON(configFilePath string) error {
	log.Printf("Looking for JSON config file in: %s\n", configFilePath)

	contents, err := ioutil.ReadFile(configFilePath)
	if err == nil {
		reader := bytes.NewBuffer(contents)
		err = json.NewDecoder(reader).Decode(&Cfg)
	}
	if err != nil {
		log.Printf("Reading configuration from JSON (%s) failed: %s\n", configFilePath, err)
	}

	return err
}

//readConfigFromENV read configuration from environment if conf.json is missing
func readConfigFromENV() (err error) {
	log.Println("Looking for ENV configuration")

	err = envconfig.Process("cm", &Cfg)
	if err == nil && Cfg.isMissing() {
		err = errors.New("Configuration is missing")
	}

	return err
}
