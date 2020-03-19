
#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

../cf-test-spring-music.sh azure-mysql small
../cf-test-spring-music.sh azure-redis small
../cf-test-spring-music.sh azure-mongodb small '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'
../cf-test-spring-music.sh azure-mssql small

