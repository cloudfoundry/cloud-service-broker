#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# bogus location or resource_group should fail
ALL_SERVICES=("csb-azure-mysql" "csb-azure-redis" "csb-azure-mongodb" "csb-azure-mssql" "csb-azure-mssql-failover-group" "csb-azure-eventhubs" "csb-azure-mssql-server" "csb-azure-postgresql")

for s in ${ALL_SERVICES[@]}; do
    ${SCRIPT_DIR}/../cf-create-service-should-fail.sh "${s}" medium '{"location":"bogus"}'
    ${SCRIPT_DIR}/../cf-create-service-should-fail.sh "${s}" medium '{"resource_group":"bogus"}'
    ${SCRIPT_DIR}/../cf-create-service-should-fail.sh "${s}" medium '{"azure_subscription_id":"bogus"}'
done

${SCRIPT_DIR}/../cf-create-service-should-fail.sh csb-azure-postgresql medium '{"sku_name":"bogus"}'

echo "$0 SUCCEEDED"