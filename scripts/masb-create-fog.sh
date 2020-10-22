#!/usr/bin/env bash

set -o errexit
set -o pipefail

if [ $# -lt 3 ]; then
    echo "Usage: ${0} <resource group> <primary server name> <secondary server name>"
    exit 1
fi

RG=$1; shift
SERVER_NAME=$1; shift
SECONDARY_SERVER_NAME=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../acceptance-tests/functions.sh"

set -o nounset

PRIMARY_LOCATION=westus
# SECONDARY_LOCATION=eastus

DB_NAME=testdb

DB_INSTANCE_NAME=masb-mssql-$$

DB_CONFIG="{
    \"resourceGroup\":\"${RG}\",
    \"location\":\"${PRIMARY_LOCATION}\",
    \"sqlServerName\":\"${SERVER_NAME}\",    
    \"sqldbName\":\"${DB_NAME}\"
}"


create_service azure-sqldb StandardS0 "${DB_INSTANCE_NAME}" "${DB_CONFIG}"


FOG_NAME=fog-$$

FOG_CONFIG="{
  \"primaryServerName\": \"${SERVER_NAME}\",
  \"primaryDbName\": \"${DB_NAME}\",
  \"secondaryServerName\": \"${SECONDARY_SERVER_NAME}\",
  \"failoverGroupName\": \"${FOG_NAME}\",
  \"readWriteEndpoint\": {
    \"failoverPolicy\": \"Automatic\",
    \"failoverWithDataLossGracePeriodMinutes\": 60
  }    
}"

create_service azure-sqldb-failover-group SecondaryDatabaseWithFailoverGroup "${FOG_NAME}" "${FOG_CONFIG}"

