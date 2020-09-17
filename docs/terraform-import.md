# Importing services from legacy service brokers
This page describes strategies for importing from legacy service brokers to the cloud service broker.  Services created by deprecated service brokers should continue to function as-is.  The cloud service broker can be installed side-by-side with legacy service brokers.  Importing to the cloud service broker is recommended when removing dependencies on deprecated service brokers is a priority or some breaking change causes a deprecated service broker to stop functioning.

> **_WARNING:_** Some migration strategies may be destructive to resources and data.  Please read and understand the processes before implementation.  If possible, test these strategies on test services and apps before applying them to important services and apps.

## Importing steps
The following steps are to import services off of a legacy service broker.
* Create new service yml file in aws broker pak and add it to manifest.yml.
* Copied over existing postgres to subsume service.
* Remove properties under plan , Plan inputs  example from yml file
* Create terraform files.
* Copy terraform provider block to subsume
* cf-fe4948ca-962d-4bc0-8825-055d7ab37c0a
* cf create-service csb-aws-postgresql-subsume small legacy-postgres-3 -c '{"aws_db_id":"cf-2edbd7fa-a53d-4bf5-a463-f42ecc679bee","admin_password":"masterpassword","region":"us-west-2"}'
* cf bind
* cf purge-service-instance from old-broker
* cf delete-service-instance csb-service-broker

