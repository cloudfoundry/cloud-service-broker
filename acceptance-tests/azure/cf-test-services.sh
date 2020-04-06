#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

cf marketplace

SPRING_MUSIC_TEST=../cf-test-spring-music.sh

if [ $# -gt 1 ]; then
  SPRING_MUSIC_TEST=../cf-test-spring-music-credhub.sh
fi


# validate services against spring music 
"${SPRING_MUSIC_TEST}" csb-azure-mysql small
"${SPRING_MUSIC_TEST}" csb-azure-redis small
"${SPRING_MUSIC_TEST}" csb-azure-mongodb small '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'
"${SPRING_MUSIC_TEST}" csb-azure-mssql small
"${SPRING_MUSIC_TEST}" csb-azure-mssql-failover-group small

./cf-test-mssql-db.sh

./cf-test-mssql-db-fog.sh

./cf-test-mssql-do-failover.sh






