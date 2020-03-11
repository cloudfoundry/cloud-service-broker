
#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

../../scripts/cf-test-service.sh azure-mysql small
../../scripts/cf-test-service.sh azure-redis small
../../scripts/cf-test-service.sh azure-mongodb small -c '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'

