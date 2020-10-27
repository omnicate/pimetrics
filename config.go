package main

import (
	"io/ioutil"
	"os"
	"time"

	modem "pimetrics/pkg/pi-modem"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	DEFAULT_PORT    = 8080
	DEFAULT_TIMEOUT = time.Minute
)

var (
	LabConfig     map[string]Config
	CurrentConfig Config
)

type Config struct {
	Tenant    string    `yaml:"tenant,omitempty"`
	Msisdn    string    `yaml:"msisdn"`
	Target    string    `yaml:"target"`
	AppConfig AppConfig `yaml:"config"`
}

type AppConfig struct {
	SwVersion   string            `yaml:"sw_version"`
	Port        uint              `yaml:"port,omitempty"`
	ModemConfig modem.ModemConfig `yaml:"modem"`
}

func readConfig(filePath string) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithError(err).Fatalf("Failed reading file: %s", filePath)
	}

	if err := yaml.Unmarshal(yamlFile, &LabConfig); err != nil {
		log.WithError(err).Fatalln("Failed unmarshalling config file to struct")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Fatalln("Failed getting hostname")
	}
	CurrentConfig = LabConfig[hostname]

	if CurrentConfig.AppConfig.Port == 0 {
		CurrentConfig.AppConfig.Port = DEFAULT_PORT
	}

	if CurrentConfig.AppConfig.ModemConfig.DefaultTimeout.String() == "" {
		CurrentConfig.AppConfig.ModemConfig.DefaultTimeout = DEFAULT_TIMEOUT
	}

	log.WithFields(log.Fields{
		"config": CurrentConfig,
	}).Info("Done reading config")
}
