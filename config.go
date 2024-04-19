package main

import (
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Username string `toml:"username"`
}

var config Config

func readConfig() {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	toml.Unmarshal(configData, &config)
}

func writeConfig() {
	configTOML, err := toml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(configFile, configTOML, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func validateConfig() bool {
	if _, err := os.Stat(configFile); err != nil {
		return false
	}

	readConfig()

	return config.Username != ""
}

func deleteConfig() {
	os.Remove(configFile)
}
