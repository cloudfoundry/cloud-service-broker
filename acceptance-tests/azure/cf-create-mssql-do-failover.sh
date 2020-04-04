#!/usr/bin/env bash

set -o nounset

NAME=$1; shift
FOG_NAME=$1; shift
RG=$1; shift
PRIMARY_SERVER=$1; shift
SECONDARY_SERVER=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

CONFIG="{ \
    \"fog_instance_name\":\"${FOG_NAME}\", \
    \"server_pair_name\":\"test\", \
    \"server_pairs\": { \
        \"test\": { \
            \"primary\": { \
                \"server_name\":\"${PRIMARY_SERVER}\", \
                \"resource_group\":\"${RG}\" \
            }, \
            \"secondary\": { \
                \"server_name\":\"${SECONDARY_SERVER}\", \
                \"resource_group\":\"${RG}\" \
            } \
        } \
    } \
}"

create_service csb-azure-mssql-do-failover standard "${NAME}" "${CONFIG}"

exit $?
