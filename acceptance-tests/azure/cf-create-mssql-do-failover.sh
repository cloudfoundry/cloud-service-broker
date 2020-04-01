#!/usr/bin/env bash

set -o nounset

NAME=$1; shift
RG=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

DB_NAME="db-${NAME}"

CONFIG="{ \
    \"fog_name\":\"mssql-${NAME}-p\", \
    \"fog_resource_group\":\"${RG}\", \
    \"db_name\":\"${DB_NAME}\" \
    }"

create_service csb-azure-mssql-do-failover standard "${NAME}-failed" "${CONFIG}"

exit $?
