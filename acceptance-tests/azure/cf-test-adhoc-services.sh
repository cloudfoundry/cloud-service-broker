#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

${SCRIPT_DIR}/cf-test-mssql-db.sh && ${SCRIPT_DIR}/cf-test-mssql-db-fog.sh && ${SCRIPT_DIR}/cf-test-mssql-do-failover.sh && ${SCRIPT_DIR}/cf-test-storage-account.sh && ${SCRIPT_DIR}/cf-test-cosmosdb.sh
RESULT=$?

exit ${RESULT}