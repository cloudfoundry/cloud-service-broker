#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

SPRING_MUSIC_TEST=../cf-test-spring-music.sh

if [ $# -gt 1 ]; then
  SPRING_MUSIC_TEST=../cf-test-spring-music-credhub.sh
fi

# bogus location or resource_group should fail
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-redis medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-redis medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"resource_group":"bogus"}'

# validate services against spring music 
"${SPRING_MUSIC_TEST}" csb-azure-mysql small
"${SPRING_MUSIC_TEST}" csb-azure-redis small
"${SPRING_MUSIC_TEST}" csb-azure-mongodb small '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'
"${SPRING_MUSIC_TEST}" csb-azure-mssql small
"${SPRING_MUSIC_TEST}" csb-azure-mssql-failover-group small

./cf-test-mssql-db.sh

./cf-test-mssql-db-fog.sh






