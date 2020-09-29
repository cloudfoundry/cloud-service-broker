# Subsuming a AWS service Broker Postgres Database Instance

It is possible to subsume control over an existing Postgres database instance that was created by the AWS-provided Service Broker. No data migration is necessary, the Cloud Service Broker just takes over the bind, unbind and destroy lifecycle.

> **_WARNING:_** Some import strategies may be destructive to resources and data.  Please read and understand the processes before implementation.  If possible, test these strategies on test services and apps before applying them to important services and apps.

> **_KNOWN ISSUE:_** Because of a few shortcomings of PostgreSQL, please be sure to follow the steps for importing services very closely. There is an issue where if the Legacy Service Instance and the newly provisioned CSB Service Instance are both bound to the same app, it will be impossible to unbind the Legacy Service Instance from the app. 

## Overview

The basic steps to subsume control of an existing AWS Postgres database instance are:

1. `cf create-service csb-aws-postgresql-subsume current <new service instance> -c '{...}'` to import the current state of the service instance.
2. `cf unbind-service <existing app> <postgres service instance>` to disconnect the app from the AWS service binding.
3. `cf bind-service <existing app> <new service instance>` to bind the app to the new service instance.
4. `cf purge-service-instance <aws service broker instance>` to make CF forget about the AWS managed service instance. *If you `cf delete-service <masb service instance>`, AWS Broker will delete the Postgres resource! Don't do this!*

### Postgres Server Administrator Password
The administrator password for the database cannot be automatically discovered. The user must provide it for the import to work. Get the Postgres Server administrator password. This was probably set when the server was created.

* Subsume the Database Instance
```bash 
cf create-service csb-aws-postgresql-subsume small csb-subsume-service -c '{"aws_db_id":"<db_instance_id>","admin_password":"<password>","region":"<region>"}'.
```   
## Unbind Application from Old Service Instance
To release the MASB provided binding, unbind the app from the old service instance.
```bash
cf unbind-service <application> <legacy instance>
```

## Bind Application to New (Subsumed) Service Instance
To create a new binding from the new service instance
```bash
cf bind-service <application> <new intance name>
```

## Assign User Permissions to CSB Postgres User  
Unfortunately, due to a known issue related to [user roles in Postgres](https://docs.pivotal.io/aws-services/postgres.html) if you bind the subsumed service to an existing applicaiton, it creates new user credentials. Make sure this new user has right permissions to ALTER and DROP existing objects or tables by running the following commands:

* Connect to the DB Instance Running PostgreSQL in AWS (See documentation here: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ConnectToPostgreSQLInstance.html)

* If the instance of PostgreSQL you are consuming was provisioned by the legacy [PCF Service Broker for AWS](https://docs.pivotal.io/aws-services/index.html), execute the following:

   ```bash
   # SET ROLE binding_group;
   ```


## 🚨 Purge Old Service Instance 🚨
CF will still have a record of the old service instance. Use `cf purge-service-instance` to purge that record. *Do not use `cf delete-service` as AWS Broker will delete the service instance that the CSB broker should now be in charge of!*
```bash
cf purge-service-instance <postgres instance>
```
* Once all services are imported you can delete old service broker. cf delete-service-instance csb-service-broker




