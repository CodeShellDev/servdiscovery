<img align="center" width="1048" height="512" alt="Dynamic service Discovery" src="https://github.com/CodeShellDev/secured-signal-api/raw/refs/heads/main/logo/banner.png" />

<p align="center">
dynamic service discovery, container syncing, automatic updates
</p>

<p align="center">
⚡ Real-time · 🔄 Automatic · 🚀 Easy to Deploy
</p>

<div align="center">
  <a href="https://github.com/codeshelldev/servdiscovery/releases">
    <img 
      src="https://img.shields.io/github/v/release/codeshelldev/servdiscovery?sort=semver&logo=github&label=Release" 
      alt="GitHub release"
    >
  </a>
  <a href="https://github.com/codeshelldev/servdiscovery/stargazers">
    <img 
      src="https://img.shields.io/github/stars/codeshelldev/servdiscovery?style=flat&logo=github&label=Stars" 
      alt="GitHub stars"
    >
  </a>
  <a href="https://github.com/codeshelldev/servdiscovery/pkgs/container/servdiscovery">
    <img 
      src="https://ghcr-badge.egpl.dev/codeshelldev/servdiscovery/size?color=%2344cc11&tag=latest&label=Image+Size&trim="
      alt="Docker image size"
    >
  </a>
  <a href="https://github.com/codeshelldev/servdiscovery/pkgs/container/servdiscovery">
    <img 
      src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fghcr-badge.elias.eu.org%2Fapi%2Fcodeshelldev%2Fservdiscovery%2Fservdiscovery&query=downloadCount&label=Downloads&color=2344cc11"
      alt="Docker image Pulls"
    >
  </a>
  <a href="./LICENSE">
    <img 
      src="https://img.shields.io/badge/License-MIT-green.svg"
      alt="License: MIT"
    >
  </a>
</div>

## Installation

> [!IMPORTANT]
> ServDiscovery works **only with Traefik**. It will **not** work with other reverse proxies due to using traefik labels to determine routes.

Get the latest `docker-compose.yaml`:

```yaml
services:
  discovery:
    image: ghcr.io/codeshelldev/servdiscovery:latest
    container_name: service-discovery
    environment:
      ENDPOINT: https://mydomain.com/discover
      ENDPOINT_KEY: MY_VERY_SECURE_KEY
      DISCOVERY_INTERVAL: 60
      ALIVE_INTERVAL: 60
      SERVER_NAME: server-1
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

Then spin it up:

```bash
docker compose up -d
```

Your discovery service is now live! 🎉

## Usage

Let's take a simple `whoami` container as an example:

```yaml
services:
  whoami:
    image: traefik/whoami:latest
    container_name: whoami
    labels:
      - traefik.enable=true
      - traefik.http.routers.whoami.rule=Host(`whoami.mydomain.com`)
      - traefik.http.routers.whoami.entrypoints=websecure
      - traefik.http.routers.whoami.tls=true
      - traefik.http.routers.whoami.tls.certresolver=cloudflare
      - traefik.http.routers.whoami.service=whoami-svc
      - traefik.http.services.whoami-svc.loadbalancer.server.port=80
      # Enable Discovery for this Container
      - discovery.enable=true
    networks:
      - traefik

networks:
  traefik:
    external: true
```

Whenever a new **Host-Rule** is added or updated, ServDiscovery will **automatically notify the configured endpoint**.  
This ensures the endpoint can correctly route traffic based on **SNI / Hostnames**.

## Endpoint Integration

ServDiscovery communicates with your endpoint via **JSON HTTP Requests**:

```json
{
	"serverName": "server-1",
	"diff": {
		"added": [
			"whoami.mydomain.com",
			"website.mydomain.com",
			"auth.mydomain.com"
		],
		"removed": [
			"whoami-backup.mydomain.com",
			"website-backup.mydomain.com",
			"auth-backup.mydomain.com"
		]
	}
}
```

Example explanation:

| ✅ Available         | ❌ Unavailable              |
| -------------------- | --------------------------- |
| whoami.mydomain.com  | whoami-backup.mydomain.com  |
| website.mydomain.com | website-backup.mydomain.com |
| auth.mydomain.com    | auth-backup.mydomain.com    |

This allows the endpoint (e.g., a load balancer) to remove `\*-backup` records from your registry and **update routable containers/services automatically**.

### Integrations

You can find example integrations inside of [examples/](./examples).

## Configuration

### `ENDPOINT_KEY`

The endpoint key is used in the `Authorization` header (Bearer token) when ServDiscovery sends POST requests.  
If no key is provided, the header is omitted.

### `DISCOVERY_INTERVAL`

Time (in seconds) between updates to your endpoint.  
**Default:** `60` seconds

### `ALIVE_INTERVAL`

Time (in seconds) between full alive discoveries. ServDiscovery sends a **complete update** of all active containers in the `added` JSON key.  
**Default:** `120` seconds

## Contributing

Found a bug or have a brilliant idea? Contributions are welcome! Open an **issue** or create a **pull request** — your help makes this project better.

## License

This project is licensed under the [MIT License](./LICENSE).
