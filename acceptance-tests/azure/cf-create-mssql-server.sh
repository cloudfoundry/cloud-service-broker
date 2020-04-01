#!/usr/bin/env bash

set -o nounset

NAME=$1; shift
USERNAME=$1; shift
PASSWORD=$1; shift
RG=$1; shift
LOCATION=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

CONFIG="{ \
    \"instance_name\":\"${NAME}\", \
    \"admin_username\":\"${USERNAME}\", \
    \"admin_password\":\"${PASSWORD}\", \
    \"resource_group\":\"${RG}\", \
    \"location\":\"${LOCATION}\" \
    }"

create_service csb-azure-mssql-server standard "${NAME}" "${CONFIG}"

exit $?