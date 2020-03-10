
#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

../../scripts/cf-test-service.sh azure-mysql small
../../scripts/cf-test-service.sh azure-redis small

