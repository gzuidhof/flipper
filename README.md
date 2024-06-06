
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./view/static/assets/logo-white.svg">
  <img width="210px" src="./view/static/assets/logo-black.svg">
</picture>

<br>
<br>
<br>


**Flipper** is a monitoring service that makes sure re-assignable IPs are always pointed at healthy hosts.

<br>

# Install
```shell
go install github.com/gzuidhof/flipper@latest
```

# Configure

First you will need to create a configuration file. You can use `yaml`, `toml` or `json`.

**`flipper.yaml`**
```yaml
# Configuration compatibility version number. Should always be 1.
version: 1

# Server is a http interface exposed by flipper.
# It is very barebones right now - one day it may be useful as a status page.
server: 
  enabled: false
  assets: 
  templates: 

# Groups are independently monitored. They each consist of some set of servers and floating IPs.
groups:
  - id: "some_group_id"
    display_name: "My Group Name"
    readonly: false # In readonly `flipper` will not take actions.
    provider: "hetzner"

    # How often should we talk to Hetzner to look for changes w.r.t. the infra itself?
    poll_interval: 60s

    hetzner:
      api_token: "abc123" # Your Hetzner API token.
      project_id: 123456 # Your Hetzner's project ID (you can find it in the URL in the Hetzner dashboard).
      servers:
        # These label selectors are how you tell flipper which servers it should watch. 
        # See the label selector docs at https://docs.hetzner.cloud/#label-selector.
        label_selector: "environment=dev,role=loadbalancer,service=my-service"
      floating_ips:
        label_selector: "environment=dev,service=my-service"
    
    checks:
      - id: "some_health_check_id"
        display_name: "Some Endpoint Health Check"
        type: "https" # "http" or "https"

        # At what interval should the check be performed.
        interval: 2s 
        # How long do we wait for the HTTP endpoint to respond.
        timeout: 4s 

        # How many successive failed checks are required to mark a server as unhealthy.
        # Note that generally this means that only after `timeout * fall` time flipper will act on unhealthy resources.
        fall: 3 
        # How many successive successful checks are required to mark a server as healthy.
        rise: 2

        # Required for HTTPS TLS check: what host do we send along and expect to receive a valid TLS cert for?
        # This is optional for "http" checks.
        host: "example.com" 

        # Path for the HTTP request.
        path: "/some/endpoint"
        
        # The port to check, defaults to 80 for "http" and 443 for "https".
        port: 443

        # Defaults to "both", but can be "ipv4" or "ipv6".
        ip_version: "both"

# Heartbeat service: it will send a HTTP GET request to the specified URL with the given interval.
# This can be used in conjunction with a (cron) monitoring service like UptimeRobot to detect when Flipper itself
# stopped running.
heartbeat:
  enabled: true
  url: "https://heartbeat.uptimerobot.com/some-url"
  interval: 60s
  timeout: 10s

telemetry:
  logging:
    level: "info" # "debug", "info", "warn", "error".
    format: "json" # "json" or "text".

notifications:
  enabled: true
  targets:
    - type: "mattermost" # Only "mattermost" is currently supported. It *probably* also works for Slack webhooks.
      channel: "my_channel"
      url: https://mattermost.example.com/hooks/some-webhook-url
      
```

# Run

```shell
# Print help text.
flipper help

# Start flipper.
flipper --config /path/to/flipper.yaml monitor
```

# Name
**flip** is a short-hand for a **floating ip**.

# How does it work?

Flipper is shipped as a standalone binary you can run on any server.

It polls Hetzner for the available floating IPs and the target servers periodically.

It performs health checks on an interval to all servers.

When a targeted server goes unhealthy it re-targets the floating IP to a different server. When it's healthy again, it will revert back to the original configuration.

## Floating IP targeting logic

Flipper generates a **deterministic** plan given the state of all the servers.

*The following rules define its behavior:*

* If all servers are unhealthy, all servers are to be considered healthy. Otherwise only healthy servers are considered.
* Targeting within a *location* is always preferred.  
  E.g. a floating IP in Hetzner `fsn1` will be targeted to a load balancer in Hetzner `fsn1` if there is at least one healthy server in that location.
* If there are multiple possible healthy servers the targeting is distributed evenly:
  * The `resource_index` label (see below) is first used within the same location to determine the canonical target.
  * Any floating IPs that are still unassigned afer the `resource_index` matching are assigned round-robin.
    The servers are ordered by the the amount of floating IPs that already target it, and then by their name.

These rules ensure that the amount of re-targets that happen is minimal to get to a even distribution.

### `resource_index`
There is a special label `resource_index` you can set on Floating IPs and servers to indicate the canonical mapping within the same location.

This is positive number that is used to minimize the amount of re-targeting when a server becomes healthy or unhealthy.

#### Why is this useful?
Imagine the following situation without `resource_index`:
```
🔥 server-0 unhealthy 
✅ server-1 healthy     <- floating-ip-0, floating-ip-3
✅ server-2 healthy     <- floating-ip-1
✅ server-3 healthy     <- floating-ip-2
```

Now `server-0` becomes healthy again. The new mapping will be

```
✅ server-0 unhealthy   <- floating-ip-0
✅ server-1 healthy     <- floating-ip-1
✅ server-2 healthy     <- floating-ip-2
✅ server-3 healthy     <- floating-ip-3
```

All the floating IPs changed!

By setting a `resource_index` label on the resources the diff be a lot smaller. This is the *before* situation with `resource_index` set.
```
🔥 server-0 unhealthy 
✅ server-1 healthy     <- floating-ip-0, floating-ip-1
✅ server-2 healthy     <- floating-ip-2
✅ server-3 healthy     <- floating-ip-3
```

### Pending server health
By default Flipper waits until it knows the health state of all servers before executing any plan.

## Supported cloud providers

It currently only supports **Hetzner** floating IPs and servers. This should be fairly easy to expand in the future.


## Tips

* It's a good idea to run Flipper in a different availability region than the resources it is watching.
* Add a label like `flipper=active` to the `label_selectors` to allow yourself to toggle off checks and switching for resources.

# Development

## Lint
You can run the linter by installing [`golanci-lint`](https://golangci-lint.run/usage/install/) and running
```shell
golangci-lint run
```

## Release
Releases are built using goreleaser, see the [goreleaser.yml](./goreleaser.yml) file.

To mint a (test) release locally, install goreleaser and run
```shell
goreleaser --snapshot --skip=publish --clean
```

# License
[MIT](./LICENSE)
