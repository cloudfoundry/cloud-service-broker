#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

# bogus region or resource_group should fail
../cf-create-service-should-fail.sh azure-mysql medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-mysql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-redis medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-redis medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-mongodb medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-mongodb medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-mssql medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-mssql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-mssql-failover-group medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-mssql-failover-group medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-eventhubs medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-eventhubs medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh azure-mssql-server standard '{"region":"bogus"}'
../cf-create-service-should-fail.sh azure-mssql-server standard '{"resource_group":"bogus"}'

# validate services against spring music 
../cf-test-spring-music.sh azure-mysql small
../cf-test-spring-music.sh azure-redis small
../cf-test-spring-music.sh azure-mongodb small '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'
../cf-test-spring-music.sh azure-mssql small
../cf-test-spring-music.sh azure-mssql-failover-group small

./cf-test-mssql-db.sh






