package config

import (
	"os"
	"strconv"

	"github.com/codeshelldev/gotl/pkg/logger"
	"github.com/codeshelldev/servdiscovery/internals/config/structure"
)

var ENV = &structure.ENV{
	DISCOVERY_INTERVAL: 60,
	ALIVE_INTERVAL: 120,
}

func Load() {
	ENV.LOG_LEVEL = os.Getenv("LOG_LEVEL")

	discoveryIntervalStr := os.Getenv("DISCOVERY_INTERVAL")

	if discoveryIntervalStr != "" {
		discoveryInterval, err := strconv.Itoa(discoveryIntervalStr)
		if err != nil {
			ENV.DISCOVERY_INTERVAL = discoveryInterval
		}
	}

	aliveIntervalStr := os.Getenv("ALIVE_INTERVAL")

	if aliveIntervalStr != "" {
		aliveInterval, err := strconv.Itoa(aliveIntervalStr)
		if err != nil {
			ENV.ALIVE_INTERVAL = aliveInterval
		}
	}

	ENV.SERVER_NAME = os.Getenv("SERVER_NAME")
	ENV.ENDPOINT = os.Getenv("ENDPOINT")
	ENV.ENDPOINT_KEY = os.Getenv("ENDPOINT_KEY")
}

func Log() {
	logger.Dev("Loaded Environment:", ENV)
}
