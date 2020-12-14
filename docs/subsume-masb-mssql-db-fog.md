# Subsuming a Microsoft Azure Service Broker MSSQL Database Fail Over Group Instance

It is possible to subsume control over an existing MSSQL database fail over group (azure-sqldb-failover-group) instance that was created by the Microsoft Azure Service Broker (MASB). No data migration is necessary, the broker just takes over the bind, unbind and destroy lifecycle.

### Pre-Requisites
- **Azure MSSQL Server Administrator Password**
The administrator password for the database cannot be automatically discovered. The user must provide it for the import to work. Get the Azure MSSQL Server administrator password. This was probably set when the server was created.

## Overview

The basic steps to subsume control of an existing  MSSQL database fail over group instance are:

1. To import the current state of the service instance run:

      ```
     cf create-service csb-azure-mssql-db-failover-group subsume "subsumed-test-db" -c 
        '{
	"azure_primary_db_id": "PRIMARY-AZURE-DB-ID",
    "azure_secondary_db_id": "SECONDARY-AZURE-DB-ID",
    "azure_fog_id": "FOG-AZURE-ID",
	"server_pair": "FOG-SERVER-PAIR",
	"server_credential_pairs": {
		"FOG-SERVER-PAIR": {
			"admin_username": "USERNAME",
			"admin_password": "PASSWORD",
            "primary":{
    			"server_resource_group": "SERVER-RESOURCE-GROUP",
    			"server_name": "PRIMARY-SERVER-NAME"
            },
            "secondary":{
    			"server_resource_group": "SERVER-RESOURCE-GROUP",
    			"server_name": "SECONDARY0-SERVER-NAME"
            }
	     }
       }'
     ```
        
2. To disconnect the app from the MASB service binding run: 

        cf unbind-service <existing app> <masb service instance>

3. To bind the app to the newly created CSB service instance run: 

        cf bind-service <existing app> <new service instance>

4. To make CF forget about the MASB managed service instance run*: 

        cf purge-service-instance <masb service instance>

 ***If you `cf delete-service <masb service instance>`, MASB will delete the resource from Azure not CF! Don't do this!**

## Details

Some details need to be provided to the `cf create-service` call so the broker can identify which IaaS resourcez to import by providing some parameters that the broker can't detect from the platform.

To subsume control of an existing MASB-brokered MSSQL database fail over group instance to
the Cloud Service Broker:

### 🔎 Azure MSSQL Resource IDs

Three ID's are necessary - the fail over group resource and the primary and secondary DB ids.

1. Get the MASB-brokered fail over group instance details by running:

    ```
    cf service MASB-SERVICE-INSTANCE
    ```

    Where is `MASB-SERVICE-INSTANCE` is the name of the service instance,
    the MAS-brokered database instance.

    For example:
```
cf service masb-fog-db-80124
Showing info of service masb-fog-db-80124 in org pivotal / space ernie as admin...

name:            masb-fog-db-80124
service:         azure-sqldb-failover-group
tags:
plan:            SecondaryDatabaseWithFailoverGroup
description:     Azure SQL Database Failover Group Service
documentation:
dashboard:

Showing status of last operation from service masb-fog-db-80124...

status:    create succeeded
message:   Created failover group masb-fog-db-80124.
started:   2020-11-11T00:39:40Z
updated:   2020-11-11T00:41:44Z

There are no bound apps for this service.
```

1. Get the Azure resource IDs for the fail over group and it's databases by running the following on the Azure CLI:

    Primary database id:
    ```
    az sql failover-group show --name FOG-NAME --server PRIMARY-SERVER-NAME --resource-group FOG-RESOURCE-GROUP --query databases[0] -o tsv
    ```
    Secondary database id:
    ```
    az sql failover-group show --name FOG-NAME --server SECONDARY-SERVER-NAME --resource-group FOG-RESOURCE-GROUP --query databases[0] -o tsv
    ```
    Fail over group id:
    ```
    az sql failover-group show --name FOG-NAME --server PRIMARY-SERVER-NAME --resource-group FOG-RESOURCE-GROUP --query id -o tsv
    ```

    Where:
    * `FOG-NAME` is the logical name in the `message` line of the output in the step above.
    * `FOG-RESOURCE-GROUP` is the **Default Resource Group** in the Default Parameters Config pane  of the Azure Service Broker tile.
      See [Default Parameters Config](https://docs.pivotal.io/partners/azure-sb/installing.html#defaultparameters-config)
      in _Installing and Configuring Microsoft Azure Service Broker_.
    * `PRIMARY-SERVER-NAME` and `SECONDARY-SERVER-NAME` are the server names joined to the fail over group, it may be necessary to consult the Azure portal to find these
 

    For example:

    <pre class="terminal">
    $ az sql failover-group show  --name fog-name --server server-name \
      --resource-group azure-service-broker --query id -o tsv
    </pre>

    The variable in the next step, `AZURE-RESOURCE-ID`, refers to the output of this command in full. An example output has been provided below 
    
    ```
    /subscriptions/000abc00-0000-00ab-0abc-00000000/resourceGroups/resource-group/providers/Microsoft.Sql/servers/server-name/databases/database-name
    ```

### 💼  Subsume the service instance

3. Create a new MSSQL service instance using Cloud Service Broker  and
   import the existing MSSQL fail over group resource by choosing the "subsume" plan and including the metadata as shown:

      ```
     cf create-service csb-azure-mssql-db-failover-group subsume NEW-SERVICE-INSTANCE -c 
        '{
	"azure_primary_db_id": "PRIMARY-AZURE-DB-ID",
    "azure_secondary_db_id": "SECONDARY-AZURE-DB-ID",
    "azure_fog_id": "FOG-AZURE-ID",
	"server_pair": "FOG-SERVER-PAIR",
	"server_credential_pairs": {
		"FOG-SERVER-PAIR": {
			"admin_username": "SERVER-ADMIN-PASSWORD",
			"admin_password": "SERVER-ADMIN-PASSWORD",
            "primary":{
    			"server_resource_group": "SERVER-RESOURCE-GROUP",
    			"server_name": "PRIMARY-SERVER-NAME"
            },
            "secondary":{
    			"server_resource_group": "SERVER-RESOURCE-GROUP",
    			"server_name": "SECONDARY0-SERVER-NAME"
            }
	     }
       }'
     ```

    Where:
    * `NEW-SERVICE-INSTANCE` is a name you choose for the new service instance
       that creates to replace the MSAB service instance.
    * `SERVER-ADMIN-PASSWORD` is the admin password for the database.
       See Prerequisite Section above.
    * `PRIMARY-AZURE-DB-ID` , `SECONDARY-AZURE-DB-ID` and `FOG-AZURE-ID` are found in the outputs of step 2 above. 

### 🔓  Disconnecting app from subsumed resource belonging to the Legacy Service Broker

4. If you used MASB to create the service instance you just subsumed, disconnect the app from the MASB service binding of the service instance you just imported by running:

    ```
    cf unbind-service APP-NAME MASB-SERVICE-INSTANCE
    ```

    Where:
    * `APP-NAME` is the app using the database.
    * `MASB-SERVICE-INSTANCE` is the name of the service instance,
    the MAS-brokered database instance.

```
$ cf unbind-service my-app masb-instance
```

### 🖇  Bind to new CSB Service Instance

5. Bind the app to the new service instance by running:

    ```
    cf bind-service APP-NAME NEW-SERVICE-INSTANCE
    ```

    Where `NEW-SERVICE-INSTANCE` is the name of the <%= vars.product_short %> service instance
    that you created in step 3 above.

    For example:
```
$cf bind-service my-app secondary-db
```
  Because Cloud Service Broker creates new credentials at bind time,
  this creates new binding credentials for the app.


### 🔁  Restage 

6. Restage the app

    ```
    cf restage test-spring
    ```

1. Remove record of the old MAS-brokered database instance and any child objects
   from Cloud Foundry by running:

    ```
    cf purge-service-instance MASB-SERVICE-INSTANCE
    ```

    For example:

**🚨 Warning:</strong> Do not run `cf delete-service`
because then the Microsoft Azure Service Broker deletes the Azure MSSQL fail over group resource, that is, the database.🚨**
