## upgrade-all-service-instances (cf cli plugin)

A CF-CLI plugin for upgrading all service instances in a CF foundation.

### Purpose
This tool was developed to allow operators to upgrade all service instances in a foundation, without having to navigate between orgs and spaces to discover upgradable instances.

### Installing
First build the binary from the plugin directory
```azure
go build .
```
Then install the plugin using the cf cli
```azure
cf install-plugin <path_to_plugin_binary>
```

### Usage

```azure
cf upgrade-all-service-instances <broker_name>

Options:
    -batch-size int "number of concurrent upgrades (default 10)"
```
