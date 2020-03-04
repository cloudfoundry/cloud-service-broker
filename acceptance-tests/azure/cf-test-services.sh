
#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

../../scripts/cf-test-service.sh azure-mysql Small
../../scripts/cf-test-service.sh azure-redis STANDARD
../../scripts/cf-test-service.sh azure-mssql Small
