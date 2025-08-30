# ServDiscovery

ServDiscovery is a Discovery Service that keeps an Endpoint updated with active Hosts (of Services).

## Installation

> [!NOTE]
> ServDiscovery only works with Traefik and not with **any** other Reverse Proxy due to `discover.enable` lable

Get the latest `docker-compose.yaml` file:

```yaml
{ { file.docker-compose.yaml } }
```

```bash
docker compose up -d
```

## Usage

Take this little `whoami` Container as an Example:

```yaml
{ { file.examples/whoami.docker-compose.yaml } }
```

Whenever a new **Host-Rule** gets added / modified ServDiscovery will update the set Endpoint to notify of any new changes.
This way the Endpoint can correctly route to different Hosts based on **SNI / Hostnames**.

## Endpoint

ServDiscovery sends requests to the Endpoint as a **JSON HTTP Request**:

```json
{ { file.examples/payload.json } }
```

This example tells the Endpoint that...

| Available            | Unavailable                 |
| -------------------- | --------------------------- |
| whoami.mydomain.com  | whoami-backup.mydomain.com  |
| website.mydomain.com | website-backup.mydomain.com |
| auth.mydomain.com    | auth-backup.mydomain.com    |

... is (un)available

This way (if the Endpoint is used by a LoadBalancer) the Owner of the Endpoint can now delete the `*-backup.mydomain.com` records from a Registry,
thus updating the list of routable Containers / Services.

## Configuration

### ENDPOINT_KEY

The Endpoint Key is provided in the Authorization Header (via Bearer) during the POST request between the Endpoint and ServDiscovery.
If no Key is provided ServDiscovery will leave out the Authorization Header.

### DISCOVERY_INTERVAL

The Discovery Interval sets the Interval (in seconds) of which ServDiscovery will update the provided Endpoint.

Default: `60`

### FULL_DISCOVERY_INTERVAL

Sets the Interval for Full Disoveries.
This tells ServDiscovery to send a Full Update of all the activate Containers in the `added` JSON Key.

> [!IMPORTANT]
> Must be set to a **fraction of** DISCOVERY_INTERVAL.
> (example: `DISCOVERY_INTERVAL * 2`)
> `FULL_DISCOVERY_INTERVAL / DISCOVERY_INTERVAL` must result in an int.
> `FULL_DISCOVERY_INTERVAL` must be bigger than `DISCOVERY_INTERVAL`.

Default: **Disabled**

## Contributing

Found a bug or have new ideas or enhancements for this Project?
Feel free to open up an issue or create a Pull Request!

## License

[MIT](https://choosealicense.com/licenses/mit/)
