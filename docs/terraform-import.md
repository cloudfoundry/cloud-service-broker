# Importing services from legacy service brokers
This page describes strategies for importing from legacy service brokers to the cloud service broker.  Services created by deprecated service brokers should continue to function as-is.  The cloud service broker can be installed side-by-side with legacy service brokers.  Importing to the cloud service broker is recommended when removing dependencies on deprecated service brokers is a priority or some breaking change causes a deprecated service broker to stop functioning.

> **_WARNING:_** Some import strategies may be destructive to resources and data.  Please read and understand the processes before implementation.  If possible, test these strategies on test services and apps before applying them to important services and apps.

### In this example we will import postgres on AWS.

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
* For importing postgres service you need aws_db_id, region and admin password for the instance that was created by old service broker.
* Import the service as below
    ```bash 
    cf create-service csb-aws-postgresql-subsume small csb-subsume-service -c '{"aws_db_id":"<db_instance_id>","admin_password":"<password>","region":"<region>"}'.
    ```
* To update the service, add the fields under
 [import_parameter_mappings](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/aws-postgres-subsume.yml#L36), [user_inputs](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/aws-postgres-subsume.yml#L56) and in [variables.tf](https://github.com/pivotal/cloud-service-broker/blob/master/aws-brokerpak/terraform/subsume-vmware-aws-postgres/postgres-variables.tf#L2). 
   
* You can unbind the old service broker instance and bind the new imported service to the application.
* To remove broker metadata related to the legacy service without deleting the underlying service's resources use [purge service instance](https://cli.cloudfoundry.org/en-US/cf/purge-service-instance.html).  Note: this action is destructive to the metadata and you will not be able to reconnect the service instance with the legacy service broker.
	```bash
	cf purge-service-instance $OLD_SERVICE_NAME
	```  
* Once all services are imported you can delete old service broker. cf delete-service-instance csb-service-broker

