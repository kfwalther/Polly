package config

import (
	"encoding/json"
	"log"
	"os"
)

// Definition of the configuration struct to house our config values for the program.
type Configuration struct {
	GcpCredentialsFile   string
	AuthTokenFile        string
	GoogleSheetsIdsFile  string
	MongoDbConnectionUri string
	MongoDbName          string
	WebServerPort        string
}

// Constructor to create a new config object from the JSON config file.
func NewConfiguration(configFile string) *Configuration {
	var c Configuration
	// Attempt to decode the JSON config and save the values.
	file, _ := os.Open(configFile)
	defer file.Close()
	jsonDecoder := json.NewDecoder(file)
	err := jsonDecoder.Decode(&c)
	if err != nil {
		log.Fatalf("Unable to parse configuration file %s: %v", configFile, err)
	}
	return &c
}
