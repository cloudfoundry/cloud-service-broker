# Subsuming a Microsoft Azure Service Broker MSSQL Database Instance

It is possible to subsume control over an existing MSSQL database instance that was created by the Microsoft Azure Service Broker (MASB). No data migration is necessary, the broker just takes over the bind, unbind and destroy lifecycle.

## Overview

The basic steps to subsume control of an existing MSSQL database instance are:

1. `cf create-service csb-masb-mssql-db-subsume current <new service instance> -c '{...}'` to import the current state of the service instance.
2. `cf unbind-service <existing app> <masb service instance>` to disconnect the app from the MASB service binding.
3. `cf bind-service <existing app> <new service instance>` to bind the app to the new service instance.
4. `cf purge-service-instance <masb service instance>` to make CF forget about the MASB managed service instance. *If you `cf delete-service <masb service instance>`, MASB will delete the Azure MSSQL resource! Don't do this!*

## Details

Some details need to be provided to the `cf create-service` call so the broker can identify which IaaS resource to import and provide some parameters that the broker can't detect from the platform.

### Azure MSSQL Resource ID

The broker needs the Azure resource ID for the database instance. Issue the following commands to get this value:

1. Get the MASB database instance details:
    ```bash
    $ cf service <masb instance>
    Showing info of service <masb instance> in org ...

    name:            <masb instance>
    service:         azure-sqldb
    tags:            
    plan:            basic
    description:     Azure SQL Database Service
    documentation:   
    dashboard:       

    Showing status of last operation from service <masb instance>...

    status:    create succeeded
    message:   Created logical database <database name> on logical server <server name>.
    started:   2020-08-20T21:06:57Z
    updated:   2020-08-20T21:08:04Z    
    ```
2. Use the `az` CLI to get the Azure resource ID for the database using <database name> and <server name> from the output above:
    ```bash
    $ az sql db show  --name <database name> --server <server name> --resource-group <server resource group> --query id -o tsv
    ```
    This value will be used as *azure resource id* below.

### Azure MSSQL Server Administrator Password
The administrator password for the database cannot be automatically discovered. The user must provide it for the import to work. Get the Azure MSSQL Server administrator password. This was probably set when the server was created.

## Subsume the Database Instance
With the proper details, the database can be subsumed:
```bash
cf create-service csb-masb-mssql-db-subsume current <new intance name> -c '{"admin_password":"<server admin password>", "azure_db_id", "<azure resource ID>"}'
```

## Unbind Application from Old Service Instance
To release the MASB provided binding, unbind the app from the old service instance.
```bash
cf unbind-service <application> <masb instance>
```

## Bind Application to New (Subsumed) Service Instance
To create a new binding from the new service instance
```bash
cf bind-service <application> <new intance name>
```

## 🚨 Purge Old Service Instance 🚨
CF will still have a record of the old service instance. Use `cf purge-service-instance` to purge that record. *Do not use `cf delete-service` as MASB will delete the service instance that the CSB broker should now be in charge of!*
```bash
cf purge-service-instance <masb instance>
```

