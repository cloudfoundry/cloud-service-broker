#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

SPRING_MUSIC_TEST=../cf-test-spring-music.sh

if [ $# -gt 1 ]; then
  SPRING_MUSIC_TEST=../cf-test-spring-music-credhub.sh
fi

allServices=("csb-azure-mysql" "csb-azure-redis" "csb-azure-mssql" "csb-azure-mssql-failover-group")

for s in ${allServices[@]}; do
  create_service "${s}" small "${s}-$$" &
done

create_service csb-azure-mongodb small csb-azure-mongodb-$$ '{"db_name": "musicdb", "collection_name": "album", "shard_key": "_id"}'

allServices+=( "csb-azure-mongodb" )

if wait; then
  RESULT=0
  for s in ${allServices[@]}; do
    if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"; then
      echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"
    else
      RESULT=1
      echo "FAILED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"
      break
    fi
  done
fi

for s in ${allServices[@]}; do
  delete_service "${s}-$$" &
done

./cf-test-mssql-db.sh
./cf-test-mssql-db-fog.sh
./cf-test-mssql-do-failover.sh

wait

exit ${RESULT}






