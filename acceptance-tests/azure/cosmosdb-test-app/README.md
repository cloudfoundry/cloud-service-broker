# cosmosdb-test-app

```bash
cf push --no-start
cf bind-service cosmosdb-test-app <cosmos service instance>
cf restart cosmosdb-test-app
```

Success if it can connect to the instance and create a collection.


