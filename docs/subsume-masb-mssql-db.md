# Subsuming a Microsoft Azure Service Broker MSSQL Database Instance

It is possible to subsume control over an existing MSSQL database instance that was created by the Microsoft Azure Service Broker (MASB). No data migration is necessary, the broker just takes over the bind, unbind and destroy lifecycle.

### Pre-Requisites
- **Azure MSSQL Server Administrator Password**
The administrator password for the database cannot be automatically discovered. The user must provide it for the import to work. Get the Azure MSSQL Server administrator password. This was probably set when the server was created.


## Overview

The basic steps to subsume control of an existing MSSQL database instance are:

1. To import the current state of the service instance run:

      ```
     cf create-service csb-azure-mssql-db subsume "subsumed-test-db" -c 
        '{
	"azure_db_id": "AZURE-DB-ID",
	"server": "PARENT-SERVER",
	"server_credentials": {
		"PARENT-SERVER": {
			"server_name": "SERVER-NAME",
			"admin_username": "USERNAME",
			"admin_password": "PASSWORD",
			"server_resource_group": "SERVER-RESOURCE-GROUP"
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

Some details need to be provided to the `cf create-service` call so the broker can identify which IaaS resource to import by providing some parameters that the broker can't detect from the platform.

To subsume control of an existing MASB-brokered MSSQL database instance to
the Cloud Service Broker:

### 🔎 Azure MSSQL Resource ID

1. Get the MAS-brokered database instance details by running:

    ```
    cf service MASB-SERVICE-INSTANCE
    ```

    Where is `MASB-SERVICE-INSTANCE` is the name of the service instance,
    the MAS-brokered database instance.

    For example:

    <pre class="terminal">$ cf service masb-instance
    Showing info of service <masb instance> in org ...

    name:            masb-instance
    service:         azure-sqldb
    tags:
    plan:            basic
    description:     Azure SQL Database Service
    documentation:
    dashboard:

    Showing status of last operation from service masb-instance...

    status:    create succeeded
    message:   Created logical database database-name on logical server server-name.
    started:   2020-08-20T21:06:57Z
    updated:   2020-08-20T21:08:04Z
    </pre>


2. Get the Azure resource ID for the database by running the following on the Azure CLI:

    ```
    az sql db show  --name DATABASE-NAME --server SERVER-NAME \
    --resource-group SERVER-RESOURCE-GROUP --query id -o tsv
    ```

    Where:
    * `DATABASE-NAME` and `SERVER-NAME` are the logical names in the `message` line of the output in the step above.
    * `SERVER-RESOURCE-GROUP` is the **Default Resource Group** in the Default Parameters Config pane
      of the Azure Service Broker tile.
      See [Default Parameters Config](https://docs.pivotal.io/partners/azure-sb/installing.html#defaultparameters-config)
      in _Installing and Configuring Microsoft Azure Service Broker_.

    For example:

    <pre class="terminal">
    $ az sql db show  --name database-name --server server-name \
      --resource-group azure-service-broker --query id -o tsv
    </pre>

    The variable in the next step, `AZURE-RESOURCE-ID`, refers to the output of this command in full. An example output has been provided below 
    
    ```
    /subscriptions/000abc00-0000-00ab-0abc-00000000/resourceGroups/resource-group/providers/Microsoft.Sql/servers/server-name/databases/database-name
    ```

### 💼 Subsume the service instance

3. Create a new MSSQL service instance using Cloud Service Broker  and
   import the existing MSSQL resource by choosing the "subsume" plan and including the metadata as shown:

      ```
     cf create-service csb-azure-mssql-db subsume "subsumed-test-db" -c 
        '{
	"azure_db_id": "AZURE-DB-ID",
	"server": "PARENT-SERVER",
	"server_credentials": {
		"PARENT-SERVER": {
			"server_name": "SERVER-NAME",
			"admin_username": "SERVER-ADMIN-USERNAME",
			"admin_password": "SERVER-ADMIN-PASSWORD",
			"server_resource_group": "SERVER-RESOURCE-GROUP"
		    }
	     }
       }'
     ```

    Where:
    * `NEW-SERVICE-INSTANCE` is a name you choose for the new service instance
       that <%= vars.product_short %> creates to replace the MSAB service instance.
    * `SERVER-ADMIN-PASSWORD` is the admin password for the database.
       See Prerequisite Section above.
    * `AZURE-RESOURCE-ID` is found in the output of step 2 above. 


    For example:

 ``` 
$ cf create-service csb-azure-mssql-db subsume secondary-db -c \
 '{"azure_db_id":"/subscriptions/000abc00-0000-00ab-0abc-00000000/resourceGroups/broker/providers/Microsoft.Sql/servers/masb-test-server/databases/test-db"}...'
 ```

### 🔓 Disconnecting app from subsumed resource belonging to the Legacy Service Broker

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

### 🖇 Bind to new CSB Service Instance

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


### 🔁 Restage 

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
because then the Microsoft Azure Service Broker deletes the Azure MSSQL resource, that is, the database.🚨**
