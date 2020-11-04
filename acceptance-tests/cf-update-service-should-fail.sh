#!/usr/bin/env bash
set -u
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

set -o pipefail

if [ -z "$1" ]; then
    echo "No service name argument supplied"
    exit 1
fi

if [ -z "$2" ]; then
    echo "No plan argument supplied"
    exit 1
fi


if [ -z "$3" ]; then
    echo "No update field argument supplied"
    exit 1
fi

SERVICE=$1; shift
PLAN=$1; shift
#UPDATE_FIELD=$1; shift
SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$-uf"

echo "creating service ..."
cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}"

if wait_for_service "${SERVICE_INSTANCE_NAME}" "create in progress" "create succeeded"; then

    for i in "$@" 
    do
    
        UPDATE_FIELD=$1
        jsonfield='{"'${UPDATE_FIELD}'":"bogus"}'
        # echo $jsonfield
        # Shift all the parameters down by one
        echo "udpating service ..."
        # udpate service..
        
        update_status=$( cf update-service "${SERVICE_INSTANCE_NAME}" -c $jsonfield | grep FAILED)

        echo $update_status

        if [ "$update_status" != "FAILED" ]; then
            delete_service "${SERVICE_INSTANCE_NAME}"
            echo "$0 FAILED"
            exit 1
        fi

        
    done

        delete_service "${SERVICE_INSTANCE_NAME}"
        echo "$0 SUCCEEDED"
        exit 0
    

fi