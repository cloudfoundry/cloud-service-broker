#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

RG_NAME=rg-test-service-$$

if create_service csb-azure-resource-group standard "${RG_NAME}" "{\"instance_name\":\"${RG_NAME}\"}"; then

  ALL_SERVICES=("csb-azure-mysql" "csb-azure-redis" "csb-azure-mssql" "csb-azure-mssql-failover-group" "csb-azure-postgresql")
  INSTANCES=()
  UPDATE_INSTANCES=()
  for s in ${ALL_SERVICES[@]}; do
    create_service "${s}" small "${s}-$$" "{\"resource_group\":\"${RG_NAME}\"}" &
    INSTANCES+=("${s}-$$")
    UPDATE_INSTANCES+=("${s}-$$")
  done

  create_service csb-azure-mongodb small csb-azure-mongodb-$$ "{\"resource_group\":\"${RG_NAME}\", \"db_name\": \"musicdb\", \"collection_name\": \"album\", \"shard_key\": \"_id\"}" &

  INSTANCES+=("csb-azure-mongodb-$$")

  NO_TLS_SERVICES=("csb-azure-mysql" "csb-azure-postgresql")

  for s in ${NO_TLS_SERVICES[@]}; do
    create_service "${s}" small "${s}-no-tls-$$" "{\"resource_group\":\"${RG_NAME}\", \"use_tls\":false}" &
    INSTANCES+=("${s}-no-tls-$$")
  done

  if wait; then
    RESULT=0
    for s in ${INSTANCES[@]}; do
      if [ $# -gt 0 ]; then
        if "${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"; then
          echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-validate-credhub.sh ${s}"
        else
          RESULT=1
          echo "FAILED: ${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"
          break
        fi
      fi

      TEST_CMD="${SCRIPT_DIR}/../cf-run-spring-music-test.sh ${s}"

      if in_list ${s} ${UPDATE_INSTANCES}; then
        echo "Will run cf update-service test on ${s}"
        TEST_CMD="${SCRIPT_DIR}/../cf-run-spring-music-test.sh ${s} medium "
      fi

      if ${TEST_CMD}; then
        echo "SUCCEEDED: ${TEST_CMD}"
      else
        RESULT=1
        echo "FAILED: ${TEST_CMD}"
        break
      fi
    done
  else
    echo "FAILED creating one or more service instances"
  fi

  for s in ${INSTANCES[@]}; do
    delete_service "${s}" &
  done

  if [ ${RESULT} -eq 0 ]; then
    ${SCRIPT_DIR}/cf-test-mssql-db.sh && ${SCRIPT_DIR}/cf-test-mssql-db-fog.sh && ${SCRIPT_DIR}/cf-test-mssql-do-failover.sh && ${SCRIPT_DIR}/cf-test-storage-account.sh && ${SCRIPT_DIR}/cf-test-cosmosdb.sh
    RESULT=$?
  fi

  wait

  delete_service "${RG_NAME}"
else
  echo "Failed creating resource group ${RG_NAME} for test services"
fi

if [ ${RESULT} -eq 0 ]; then
  echo "SUCCEEDED: $0"
else
  echo "FAILED: $0"
fi

exit ${RESULT}






