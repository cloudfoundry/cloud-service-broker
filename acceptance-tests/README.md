# Acceptance Test Tools and Scripts

Acceptance tests are run as `cf push`'ed applications that verify connectivity in a real Cloud Foundry environment on the respective IaaS (GCP, AWS, Azure.)

The main pattern is:
1. `cf create-service` the service instance to test
2. `cf bind-service` a test app to the provisioned service
3. `cf restart` the test app and verify that it successfully starts
4. `cf delete-service` the tested instance 

## Pipeline tests

The build pipeline has acceptance jobs for each IaaS (GCP, Azure and AWS.) 

### Positive Tests
Positive tests follow the basic pattern above - create a service, bind a test, validate the test app is successful.

### Negative Tests
Negative test verify that bad configuration options result in service creation failure. Its often helpful to make sure configuration parameters are honored (as opposed to being ignored) and making sure bad values result in failures helps verify the options are honored - if the creation succeeds, it often means the option is being ignored somewhere.


| Positive Tests | Negative Tests |
|----------------|----------------|
| [aws/cf-test-services.sh](./aws/cf-test-services.sh) | [azure/cf-test-negative-responses.sh](./azure/cf-test-negative-responses.sh) |
| [azure/cf-test-services.sh](./azure/cf-test-services.sh) | [azure/cf-test-negative-responses.sh](./azure/cf-test-negative-responses.sh) |
| [gcp/cf-test-services.sh](./gcp/cf-test-services.sh) | [gcp/cf-test-negative-responses.sh](./gcp/cf-test-negative-responses.sh) |

## Spring Music

Many of the services (MySQL, PostgreSQL, SQL Server, MongoDB, Redis) are supported by [SpringMusic](https://github.com/cloudfoundry-samples/spring-music). These services are tested a little differently:
1. `cf create-service` the service instance to test
2. `cf bind-service` the SpringMusic app to the provisioned service
3. `cf restart` SpringMusic and verify that it successfully starts
4. `cf unbind-service` SpringMusic from the service instance
5. `cf update-service` (if called with -u <plan>) to make sure update does not result in data loss
6. `cf bind-service` the [spring-music-validator](./spring-music-validator) app to the provisioned service
7. `cf restart` spring-music-validator and verify it successfully starts 
8. `cf delete-service` the tested instance 

The use of the spring-music-validator app verifies all the binding credential variations that each service provides and validates that SpringMusic successfully initialed each service with its seed data.

## Other Services

For those services that SpringMusic doesn't support, some validation app that can be `cf push`'ed has been created. See [AWS](./aws) and [Azure](./azure)

When new services are added, test apps should be created and the corresponding IaaS `cf-test-services.sh` should be updated so the test gets run as part of the pipeline.