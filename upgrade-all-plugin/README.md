## upgrade-all-service-instances (cf cli plugin)

A CF-CLI plugin for upgrading all service instances in a CF foundation.

### Purpose
This tool was developed to allow users to upgrade all service instances they have access to in a foundation, without having to navigate between orgs and spaces to discover upgradable instances.

**Warning:** It is important to ensure that the authenticated user only has access to instances which you wish to upgrade. The plugin will upgrade all service instances a user has access to, irrespective of org and space. 

### Installing
First build the binary from the plugin directory
```
go build .
```
Then install the plugin using the cf cli
```
cf install-plugin <path_to_plugin_binary>
```

### Usage

```
cf upgrade-all-service-instances <broker_name>

Options:
    -batch-size int "number of concurrent upgrades (default 10)"
```
