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

	discoveryInterval, err := strconv.Atoi(os.Getenv("DISCOVERY_INTERVAL"))

	if err != nil {
		if discoveryInterval > 0 {
			ENV.DISCOVERY_INTERVAL = discoveryInterval
		}
	}

	aliveInterval, err := strconv.Atoi(os.Getenv("ALIVE_INTERVAL"))

	if err != nil {
		if aliveInterval > 0 {
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