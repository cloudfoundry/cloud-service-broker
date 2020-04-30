# spring-music-validator

A node.js app that can be `cf push`'ed to Cloud Foundry, bound to a service instance that [SpringMusic](https://github.com/cloudfoundry-samples/spring-music) has been bound to and will fail to start if the binding fails or it doesn't find data that SpringMusic should have written to the service instance.

Useful in validating all credential information for common services (MySQL, PostgreSQL, MongoDB, Redis, SQLServer.)

Used by [cf-run-spring-music-test.sh](../cf-run-spring-music-test.sh) to validate services.