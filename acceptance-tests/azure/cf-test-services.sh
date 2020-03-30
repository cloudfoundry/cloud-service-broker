#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

# bogus region or resource_group should fail
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-redis medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-redis medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"region":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"resource_group":"bogus"}'

# validate services against spring music 
../cf-test-spring-music.sh csb-azure-mysql small
../cf-test-spring-music.sh csb-azure-redis small
../cf-test-spring-music.sh csb-azure-mongodb small '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'
../cf-test-spring-music.sh csb-azure-mssql small
../cf-test-spring-music.sh csb-azure-mssql-failover-group small

./cf-test-mssql-db.sh

./cf-test-mssql-db-fog.sh






