from docker import DockerClient
from urllib.parse import urlparse as parseUrl
import requests
import os
import signal
import logging
import re
import sys
from time import sleep

# import threading

logger = logging.getLogger("ServDiscovery")

handler = logging.StreamHandler(sys.stdout)
formatter = logging.Formatter(
    fmt="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%d.%m %H:%M"
)
handler.setFormatter(formatter)
logger.addHandler(handler)

LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")


SERVER_NAME = os.getenv("SERVER_NAME")

ENDPOINT = os.getenv("ENDPOINT")
ENDPOINT_KEY = os.getenv("ENDPOINT_KEY")

DISCOVERY_INTERVAL = os.getenv("DISCOVERY_INTERVAL")
FULL_DISCOVERY_INTERVAL = os.getenv("FULL_DISCOVERY_INTERVAL")

fullDiscoveryRatio = 0

dockerClient = DockerClient(base_url="unix://var/run/docker.sock")

# Threading
# containerToHostLock = threading.Lock()
# containerLock = threading.Lock()

containersToHosts = {}
containers = {}

def getDiff(old, new):
    old_set, new_set = set(old), set(new)
    
    removed = old_set - new_set
    added = new_set - old_set

    return {
        "removed": list(removed),
        "added": list(added)
    }

def getHostsFromLabels(labels):
    hosts = []

    for key, value in labels.items():
        if key.startswith("traefik.http.routers.") and key.endswith(".rule"):
            matches = re.findall(r"Host\(`([^`]+)`\)", value)
            hosts.extend(matches)

    return hosts

def getEnabledContainers():
    newContainers = []

    try:
        newContainers = dockerClient.containers.list(filters={"label": "discovery.enable=true"})
    except Exception as e:
        logger.error(f"Error fetching containers: {str(e)}")

    return newContainers

def getDiscovery():
    diffs = {
        "removed": [],
        "added": []
    }

    global containers

    newContainers = getEnabledContainers()

    containerDiff = getDiff(containers, newContainers)

    logger.info(f"Found {newContainers.__len__()} Containers")

    logger.debug(f"Found {containerDiff.get("added").__len__()} added Containers")
    logger.debug(f"Found {containerDiff.get("removed").__len__()} removed Containers")

    # Threading
    # with containerToHostLock and containerLock:

    containers = newContainers

    # Update changed Containers and Add new Containers
    for container in newContainers:
        hosts = getHostsFromLabels(container.labels)

        if container.id in containersToHosts:
            old = containersToHosts[container.id]
            new = hosts

            # Get Difference
            diff = getDiff(old, new)

            logger.debug(f"[{container.name}] + {diff.get("added")}, - {diff.get("removed")}")

            diffs.get("removed").extend(diff.get("removed"))
            diffs.get("added").extend(diff.get("added"))

            containersToHosts[container.id] = hosts
        else:
            logger.debug(f"Added {container.name}")

            diffs.get("added").extend(hosts)
            
            containersToHosts[container.id] = hosts

    # Diff Old / Removed Containers
    for removedContainer in containerDiff.get("removed"):
        if removedContainer.id in containersToHosts:

            # Get all Hosts from Removed Container and Add them to the global Diff
            diffs.get("removed").extend(containersToHosts[removedContainer.id])
            
            # Remove Container from Dict
            logger.debug(f"Removed {removedContainer.name}")

            containersToHosts.pop(removedContainer.id)

    return diffs

def getFullDiscovery():
    diffs = {
        "removed": [],
        "added": []
    }

    newContainers = getEnabledContainers()

    logger.info(f"Found {newContainers.__len__()} Containers")

    for container in newContainers:
        hosts = getHostsFromLabels(container.labels)

        diffs.get("added").extend(hosts)

    return diffs

def cleanDiff(diff):
    removed, added = set(diff.get("removed")), set(diff.get("added"))

    both = removed & added

    removed -= both
    added -= both

    return {
        "removed": list(removed),
        "added": list(added)
    }

def exitContainer():
    logger.error(f"Shutting Container down...")

    os.kill(os.getpid(), signal.SIGTERM)

def sendDiffToEndpoint(diff):
    data =  { "serverName": SERVER_NAME, "diff": diff }

    headers = {}
    
    if ENDPOINT_KEY:
        headers["Authorization"] = f"Bearer {ENDPOINT_KEY}"

    response = requests.post(
        url=ENDPOINT,
        json=data,
        headers=headers
    )

    return response

# Threading
# def startBackgroundThread():
#   thread = threading.Thread(target=main, daemon=True)
#   thread.start()

def doDiscovery():
    logger.debug("Starting Discovery")

    globalDiff = getDiscovery()

    logger.debug("Cleaning Diff")

    return cleanDiff(globalDiff)

def doFullDiscovery():
    logger.debug("Starting Full Discovery")

    globalDiff = getFullDiscovery()

    return globalDiff

def main():
    i = 0
    while True:
        isFullDiscovery = False

        if i == fullDiscoveryRatio:
            isFullDiscovery = True
            i = 0

        if not isFullDiscovery:
            logger.info(f"Starting Discover in {DISCOVERY_INTERVAL}...")
        else:
            logger.info(f"Starting Full Discover in {DISCOVERY_INTERVAL}...")

        sleep(DISCOVERY_INTERVAL)

        globalDiff = {
            "removed": [],
            "added": []
        }

        if not isFullDiscovery:
            globalDiff = doDiscovery()
        else:
            globalDiff = doFullDiscovery()

        # Check if there are any Changes
        if globalDiff.get("removed").__len__() + globalDiff.get("added").__len__() <= 0:
            logger.debug("No Changes were made, skipping...")
            continue
        
        logger.info(f"Sending Diff to {ENDPOINT} with{"out" if not ENDPOINT_KEY else ""} Auth")

        response = sendDiffToEndpoint(globalDiff)

        logger.debug(f"Endpoint responded with {response.text} {response.status_code} {"OK" if response.ok else "NOT OK"}")

        i += 1

if __name__ == '__main__':
    logger.setLevel(level=LOG_LEVEL)

    if not SERVER_NAME or not ENDPOINT:
        if not SERVER_NAME:
            logger.error(f"No SERVER_NAME set")
        if not ENDPOINT:
            logger.error(f"No ENDPOINT set")

        exitContainer()

    if not DISCOVERY_INTERVAL:
        logger.warning(f"No DISCOVERY_INTERVAL set, using 30sec as default")
        DISCOVERY_INTERVAL = 30
    else:
        DISCOVERY_INTERVAL = int(DISCOVERY_INTERVAL)

    if not FULL_DISCOVERY_INTERVAL:
        logger.warning(f"No FULL_DISCOVERY_INTERVAL set, disregarding Full Discoveries")
        FULL_DISCOVERY_INTERVAL = 0
    else:
        FULL_DISCOVERY_INTERVAL = int(FULL_DISCOVERY_INTERVAL)

    if FULL_DISCOVERY_INTERVAL % DISCOVERY_INTERVAL != 0:
        logger.warning(f"FULL_DISCOVERY_INTERVAL is not a valid Integer, rounding to nearest Int")
        FULL_DISCOVERY_INTERVAL = round(FULL_DISCOVERY_INTERVAL)

    if FULL_DISCOVERY_INTERVAL != 0:
        fullDiscoveryRatio = round(FULL_DISCOVERY_INTERVAL / DISCOVERY_INTERVAL)
        logger.debug(f"FULL_DISCOVERY_INTERVAL ratio set to {fullDiscoveryRatio}")

    if not ENDPOINT_KEY or ENDPOINT_KEY == "":
        logger.warning(f"No ENDPOINT_KEY set, requests may be denied")
    
    main()