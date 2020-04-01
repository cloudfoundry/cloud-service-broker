#!/usr/bin/env bash

set -o nounset

NAME=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

PRIMARY_SERVER_NAME="mssql-${NAME}-p"
SECONDARY_SERVER_NAME="mssql-${NAME}-s"

${SCRIPT_DIR}/cf-delete-mssql-server.sh "${PRIMARY_SERVER_NAME}"
${SCRIPT_DIR}/cf-delete-mssql-server.sh "${SECONDARY_SERVER_NAME}" 
delete_service "${NAME}"

exit $?