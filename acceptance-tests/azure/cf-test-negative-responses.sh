#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# bogus location or resource_group should fail
ALL_SERVICES=("csb-azure-mysql" "csb-azure-redis" "csb-azure-mongodb" "csb-azure-mssql" "csb-azure-mssql-failover-group" "csb-azure-postgresql" "csb-azure-cosmosdb-sql")

BOGUS_OPTIONS=('{"location":"bogus"}' '{"resource_group":"bogus"}' '{"azure_subscription_id":"bogus"}')

for s in ${ALL_SERVICES[@]}; do
    for o in ${BOGUS_OPTIONS[@]}; do
        ${SCRIPT_DIR}/../cf-create-service-should-fail.sh "${s}" medium ${o}
    done
done

STANDARD_PLAN_SERVICES=("csb-azure-storage-account" "csb-azure-mssql-server")
for s in ${STANDARD_PLAN_SERVICES[@]}; do
    for o in ${BOGUS_OPTIONS[@]}; do
        ${SCRIPT_DIR}/../cf-create-service-should-fail.sh "${s}" standard ${o}
    done
done

for o in ${BOGUS_OPTIONS[@]}; do
    ${SCRIPT_DIR}/../cf-create-service-should-fail.sh csb-azure-eventhubs basic ${o}
done

${SCRIPT_DIR}/../cf-create-service-should-fail.sh csb-azure-postgresql medium '{"sku_name":"bogus"}'

echo "$0 SUCCEEDED"