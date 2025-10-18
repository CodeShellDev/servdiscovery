# ServDiscovery

ServDiscovery is a Discovery Service that keeps an Endpoint updated with active Hosts (of Services).

## Installation

> [!NOTE]
> ServDiscovery only works with Traefik and not with **any** other Reverse Proxy due to `discovery.enable` label

Get the latest `docker-compose.yaml` file:

```yaml
{
services:
  discovery:
    image: ghcr.io/codeshelldev/servdiscovery:latest
    container_name: service-discovery
    environment:
      ENDPOINT: https://mydomain.com/ENDPOINT
      ENDPOINT_KEY: MY_VERY_SECURE_KEY
      SERVER_NAME: server-1
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

```bash
docker compose up -d
```

## Usage

Take this little `whoami` Container as an Example:

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
      # Enable Discovery on this Container
      - discovery.enable=true
    networks:
      - traefik

networks:
  traefik:
    external: true
```

Whenever a new **Host-Rule** gets added / modified ServDiscovery will update the set Endpoint to notify of any new changes.
This way the Endpoint can correctly route to different Hosts based on **SNI / Hostnames**.

## Endpoint

ServDiscovery sends requests to the Endpoint as a **JSON HTTP Request**:

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

This example tell the Endpoint that...

| Available            | Unavailable                 |
| -------------------- | --------------------------- |
| whoami.mydomain.com  | whoami-backup.mydomain.com  |
| website.mydomain.com | website-backup.mydomain.com |
| auth.mydomain.com    | auth-backup.mydomain.com    |

This way (if the Endpoint is used by a LoadBalancer) the Owner of the Endpoint can now delete the `*-backup.mydomain.com` records from a Registry,
thus updating the list of routable Containers / Services.

## Configuration

### ENDPOINT_KEY

The Endpoint Key is provided in the Authorization Header (via Bearer) during the POST request between the Endpoint and ServDiscovery.
If no Key is provided ServDiscovery will leave out the Authorization Header.

### DISCOVERY_INTERVAL

The Discovery Interval sets the Interval of which ServDiscovery will update the Endpoint, etc.

## Contributing

Found a bug or have new ideas or enhancements for this Project?
Feel free to open up an issue or create a Pull Request!

## License

[MIT](https://choosealicense.com/licenses/mit/)
