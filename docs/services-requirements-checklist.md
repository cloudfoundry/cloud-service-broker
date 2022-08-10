# Services requirements checklist

The following is a list of functional and non-functional requirements that are generally accepted as needed for a production-ready service offering. Depending on the service, the customer requests and the timing we can decide how to prioritise them and what to include in the first GA release. 
Keep in mind that some features if introduced later might force us to make breaking changes, so it might be good to add them early on.

## Features: 
* [ ] Security by default
  * [ ] Enable or expose TLS connection in all steps
  * [ ] Enable or expose encryption at rest 
  * [ ] Use strong passwords
* [ ] High Availablity by default
* [ ] Enable or expose multi zone replication
* [ ] Backup and restore capabilities by default 
* [ ] Implement IaaS recommended features - Sometimes the IaaS recommends enabling certain features. Those can be good indications of the features that we should enable our customers to use as well. 
* [ ] Ability to add custom plans via configuration
* [ ] Ability to remove the pre-defined build-in plans
* [ ] Ability to rotate “hidden”/admin credentials if compromised

## Defaults: 
* [ ] Follow IaaS recommended defaults- An easy way to find this out is to attempt to create an instance from the console of the IaaS. For a service that has already been GAed, we might not be able to change the defaults because it would cause a breaking change.

## Supportability:
* [ ] All TF providers used are supported
* [ ] All the properties used are supported
* [ ] All dependencies are up to date
* [ ] All providers are up to date
* [ ] TF binary version is up to date

## Testing:
* [ ] Acceptance testing of the end-to-end flow including provision/update/bind/unbind/delete
* [ ] Upgrade testing that validates we are keeping backward compatibility with the previously released version
* [ ] Integration testing of the properties passed to TF
* [ ] The service plan can be updated

## Documentation:
* [ ] All properties are documented
* [ ] The bind/unbind process is documented
* [ ] There is an upgrading guide (if needed)
