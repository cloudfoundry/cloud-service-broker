#!/usr/bin/env bash

set -o nounset

NAME=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

delete_service "${NAME}"

PRIMARY_SERVER_NAME="mssql-${NAME}-p"
SECONDARY_SERVER_NAME="mssql-${NAME}-s"

if [ $# -gt 0 ]; then
  PRIMARY_SERVER_NAME=$1; shift
fi

if [ $# -gt 0 ]; then
  SECONDARY_SERVER_NAME=$1; shift
fi

${SCRIPT_DIR}/cf-delete-mssql-server.sh "${PRIMARY_SERVER_NAME}" &
${SCRIPT_DIR}/cf-delete-mssql-server.sh "${SECONDARY_SERVER_NAME}" &

wait

exit $?