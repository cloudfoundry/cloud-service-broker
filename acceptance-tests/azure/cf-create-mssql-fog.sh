#!/usr/bin/env bash

set -o nounset

NAME=$1; shift
USERNAME=$1; shift
PASSWORD=$1; shift
RG=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

DB_NAME="db-${NAME}"
PRIMARY_SERVER_NAME="mssql-${NAME}-p"
SECONDARY_SERVER_NAME="mssql-${NAME}-s"
if ${SCRIPT_DIR}/cf-create-mssql-server.sh "${PRIMARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${RG}" westus; then
    if ${SCRIPT_DIR}/cf-create-mssql-server.sh "${SECONDARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${RG}" eastus; then
        CONFIG="{
          \"instance_name\":\"${NAME}\", \
          \"db_name\":\"${DB_NAME}\", \
          \"server_pair\":\"test\", \
          \"server_credential_pairs\":{ \
            \"test\":{ \
              \"admin_username\":\"${USERNAME}\", \
              \"admin_password\":\"${PASSWORD}\", \
              \"primary\":{ \
                \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                \"resource_group\":\"${RG}\" \
              }, \
              \"secondary\":{ \
                \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                \"resource_group\":\"${RG}\" \
              } \
            } \
          } \
        }"
        if create_service csb-azure-mssql-db-failover-group medium "${NAME}" "${CONFIG}"; then
            echo "Successfully created failover group ${NAME}"
        else
            delete_service "${SECONDARY_SERVER_NAME}"
            delete_service "${PRIMARY_SERVER_NAME}"
        fi
    else
        delete_service "${PRIMARY_SERVER_NAME}"
    fi
fi

exit $?