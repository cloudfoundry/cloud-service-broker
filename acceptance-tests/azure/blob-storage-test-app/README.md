# blob-storage-test-app

```bash
cf push --no-start
cf bind-service blob-storage-test-app <storage account service instance>
cf restart blob-storage-test-app
```

Success if it can connect to the instance enumerate containers.