#!/usr/bin/env bash

set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

NAME=fog-test-$$
PRIMARY_SERVER_NAME="mssql-${NAME}-p"
SECONDARY_SERVER_NAME="mssql-${NAME}-s"
USERNAME="testadminuser"
PASSWORD="A_C0mpl1cated-Passw0rd"
RG="csb-acceptance-test-rg"
FAILOVER_NAME=fog-test-failed-$$

RESULT=1
if "${SCRIPT_DIR}/cf-create-mssql-fog.sh" "${NAME}" "${USERNAME}" "${PASSWORD}" "${RG}" "${PRIMARY_SERVER_NAME}" "${SECONDARY_SERVER_NAME}"; then
  if "${SCRIPT_DIR}/cf-create-mssql-do-failover.sh" "${FAILOVER_NAME}" "${NAME}" "${RG}" "${PRIMARY_SERVER_NAME}" "${SECONDARY_SERVER_NAME}"; then
    if delete_service "${NAME}"; then
      cf service "${NAME}"
      echo "Should not have been able to delete failover group in swapped state!"
      delete_service "${FAILOVER_NAME}"
    else
      delete_service "${FAILOVER_NAME}"
      RESULT=$?
    fi  
  fi
fi

"${SCRIPT_DIR}/cf-delete-mssql-fog.sh" "${NAME}" "${PRIMARY_SERVER_NAME}" "${SECONDARY_SERVER_NAME}"

exit ${RESULT}