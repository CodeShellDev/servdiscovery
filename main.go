package main

import (
	"sync"
	"time"

	log "github.com/codeshelldev/gotl/pkg/logger"
	"github.com/codeshelldev/servdiscovery/internals/config"
	"github.com/codeshelldev/servdiscovery/internals/discovery"
	"github.com/codeshelldev/servdiscovery/internals/docker"
)

func main() {
	config.Load()

	log.Init(config.ENV.LOG_LEVEL)

	docker.Init()

	log.Info("Initialized Logger with Level of ", log.Level())

	if log.Level() == "dev" {
		log.Dev("Welcome back Developer!")
	}

	config.Log()

	docker.InitClient()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if config.ENV.DISCOVERY_INTERVAL <= 0 {
			log.Info("Disabling diff discoveries")
			return
		}

		log.Debug("Started discovery loop")

		ticker := time.NewTicker(time.Duration(config.ENV.DISCOVERY_INTERVAL) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			process(discovery.GetDiffDiscovery())
		}
	}()

	go func() {
		defer wg.Done()
		
		if config.ENV.ALIVE_INTERVAL <= 0 {
			log.Info("Disabling alive discoveries")
			return
		}

		log.Debug("Started alive discovery loop")

		ticker := time.NewTicker(time.Duration(config.ENV.ALIVE_INTERVAL) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			process(discovery.GetAliveDiscovery())
		}
	}()

	stop := docker.Run(func() {
		wg.Wait()
	})

	<-stop
	docker.Shutdown()
}

func process(diff discovery.Diff[string]) {
	log.Dev("Received diff: ", diff)

	if len(diff.Added) <= 0 && len(diff.Removed) <= 0 {
		log.Info("No changes detected, skipping...")
		return
	}

	log.Debug("Sending diff to ", config.ENV.ENDPOINT, " with ", "TOKEN:", config.ENV.ENDPOINT_KEY)

	resp, err := discovery.SendDiff(config.ENV.SERVER_NAME, config.ENV.ENDPOINT, config.ENV.ENDPOINT_KEY, diff)

	if err != nil {
		log.Error("Error sending diff: ", err.Error())
		return
	}

	log.Debug("Endpoint responded with ", resp.Status)
}
