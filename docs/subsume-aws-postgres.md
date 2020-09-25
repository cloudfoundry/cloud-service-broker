# Subsuming a AWS service Broker Postgres Database Instance

It is possible to subsume control over an existing Postgres database instance that was created by the AWS Service Broker. No data migration is necessary, the broker just takes over the bind, unbind and destroy lifecycle.

> **_WARNING:_** Some import strategies may be destructive to resources and data.  Please read and understand the processes before implementation.  If possible, test these strategies on test services and apps before applying them to important services and apps.

> **_WARNING:_** Known issue, if you bind subsumed service to existing applicaiton, It creates new user credentials. Make sure this user has right permissions to ALTER and DROP existing objects or tabels.

## Overview

The basic steps to subsume control of an existing MSSQL database instance are:

1. `cf create-service csb-aws-postgresql-subsume current <new service instance> -c '{...}'` to import the current state of the service instance.
2. `cf unbind-service <existing app> <postgres service instance>` to disconnect the app from the AWS service binding.
3. `cf bind-service <existing app> <new service instance>` to bind the app to the new service instance.
4. `cf purge-service-instance <aws service broker instance>` to make CF forget about the AWS managed service instance. *If you `cf delete-service <masb service instance>`, AWS Broker will delete the Postgres resource! Don't do this!*

## Importing steps
The following steps are followed to import services of a legacy service broker.
* Create a new subsume service yml file in AWS broker pak and add it to the manifest.yml. [Sample subsume service](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/aws-postgres-subsume.yml)
* Create terraform these terraform files.
    * ``provider.tf``
    * ``main.tf``
    * ``variables.tf``
    * ``outputs.tf``
* Update ``provider.tf ``with corresponding provider.
* Copy the resource block for the serivce you want to import.
* Create admin password variable in ``varialbes.tf``. This value is used for creating new credentials when binding the app.
* Build and deploy brokerpak. 
* To update the service, add the fields under
 [import_parameter_mappings](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/aws-postgres-subsume.yml#L36), [user_inputs](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/aws-postgres-subsume.yml#L56) and in [variables.tf](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/terraform/subsume-vmware-aws-postgres/postgres-variables.tf#L2). 
### Postgres Server Administrator Password
The administrator password for the database cannot be automatically discovered. The user must provide it for the import to work. Get the Postgres Server administrator password. This was probably set when the server was created.

* Subsume the Database Instance
    ```bash 
    cf create-service csb-aws-postgresql-subsume small csb-subsume-service -c '{"aws_db_id":"<db_instance_id>","admin_password":"<password>","region":"<region>"}'.
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
Known issue arround user postgres [roles](https://docs.pivotal.io/aws-services/postgres.html)

## 🚨 Purge Old Service Instance 🚨
CF will still have a record of the old service instance. Use `cf purge-service-instance` to purge that record. *Do not use `cf delete-service` as AWS Broker will delete the service instance that the CSB broker should now be in charge of!*
```bash
cf purge-service-instance <postgres instance>
```
* Once all services are imported you can delete old service broker. cf delete-service-instance csb-service-broker




