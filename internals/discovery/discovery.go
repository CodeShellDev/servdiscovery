package discovery

import (
	"bytes"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/codeshelldev/gotl/pkg/jsonutils"
	"github.com/codeshelldev/gotl/pkg/logger"
	"github.com/codeshelldev/servdiscovery/internals/docker"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

var containerHosts = map[string][]string{}

var containers []container.Summary

func GetDiffDiscovery() Diff[string] {
	logger.Debug("Starting discovery")

	diff, err := getContainerDiff()

	if err != nil {
		logger.Error("Encountered error during discovery: ", err.Error())
		return Diff[string]{}
	}

	logger.Debug("Cleaning diff")

	cleaned := CleanDiff(diff)

	return cleaned
}

func GetAliveDiscovery() Diff[string] {
	logger.Debug("Starting alive discovery")

	globalDiff := Diff[string]{
		Added: []string{},
		Removed: []string{},
	}

	newContainers, err := getEnabledContainers()

	if err != nil {
		logger.Error("Encountered error during discovery: ", err.Error())
		return Diff[string]{}
	}

	logger.Info("Found ", len(newContainers), " enabled containers")

	for _, container := range newContainers {
		router := getRouterHosts(container)
		
		seq := maps.Values(router)
		hostSlices := slices.Collect(seq)
		hosts := slices.Concat(hostSlices...)

		globalDiff.Added = append(globalDiff.Added, hosts...)
	}

	cleaned := CleanDiff(globalDiff)

	return cleaned
}

func SendDiff(serverName, endpoint, key string, diff Diff[string]) (*http.Response, error) {
	payload := map[string]any{
		"serverName": serverName,
		"diff": diff,
	}

	data, err := jsonutils.ToJsonSafe(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if key != "" {
		req.Header.Set("Authorization", key)
	}

	resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
	defer resp.Body.Close()

	return resp, nil
}

func getContainerDiff() (Diff[string], error) {
	globalDiff := Diff[string]{
		Added: []string{},
		Removed: []string{},
	}

	newContainers, err := getEnabledContainers()

	if err != nil {
		return Diff[string]{}, err
	}

	containerDiff := diffContainers(containers, newContainers)

	containers = newContainers

	logger.Info("Found ", len(containers), " enabled containers")
	
	logger.Debug("Found ", len(containerDiff.Added), " added containers")
	logger.Debug("Found ", len(containerDiff.Removed), " removed containers")

	for _, container := range newContainers {
		router := getRouterHosts(container)
		
		seq := maps.Values(router)
		hostSlices := slices.Collect(seq)
		hosts := slices.Concat(hostSlices...)

		old, exists := containerHosts[container.ID]
		if exists {
			diff := GetDiff(old, hosts)

			logDiff(container.Names[0], diff)

			globalDiff.Merge(diff)
		} else {
			logger.Info("Added ", container.Names[0])
		}

		containerHosts[container.ID] = hosts 
	}

	for _, removed := range containerDiff.Removed {
		host, exists := containerHosts[removed.ID]

		if exists {
			globalDiff.Removed = append(globalDiff.Removed, host...)

			logger.Info("Removed ", removed.Names[0])

			delete(containerHosts, removed.ID)
		}
	}

	return globalDiff, nil
}

func diffContainers(old, new []container.Summary) Diff[container.Summary] {
    oldIDs := make([]string, 0, len(old))
    newIDs := make([]string, 0, len(new))

	oldContainers := map[string]container.Summary{}
	newContainers := map[string]container.Summary{}

    for _, container := range old {
        oldIDs = append(oldIDs, container.ID)
		oldContainers[container.ID] = container
    }
    for _, container := range new {
        newIDs = append(newIDs, container.ID)
		newContainers[container.ID] = container
    }

    idDiff := GetDiff(oldIDs, newIDs)

	var diff Diff[container.Summary]

	for _, added := range idDiff.Added {
		diff.Added = append(diff.Added, newContainers[added])
	}
	for _, removed := range idDiff.Removed {
		diff.Removed = append(diff.Removed, oldContainers[removed])
	}

	return diff
}

func getRouterHosts(container container.Summary) map[string][]string {
	hosts := map[string][]string{}

	hostRegex, err := regexp.Compile(`Host\(\x60([^\x60]+)\x60\)`)

	if err != nil {
		return nil
	}

	routerRegex, err := regexp.Compile(`traefik\.http\.routers\.([A-Za-z0-9._-]+)\.rule`)

	if err != nil {
		return nil
	}

	for key, value := range container.Labels {
		router := routerRegex.FindString(key)
		if router != "" {
			hosts[router] = hostRegex.FindStringSubmatch(value)
		}
	}

	return hosts
}

func getEnabledContainers() ([]container.Summary, error) {
	filters := client.Filters{}
	filters.Add("label", "discovery.enable=true")

	return docker.GetContainers(client.ContainerListOptions{
		Filters: filters,
	})
}