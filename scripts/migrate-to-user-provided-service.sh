#!/usr/bin/env bash

set -o errexit
set -o pipefail

if [ -z "$1" ]; then
    echo "No app name argument supplied"
    exit 1
fi

if [ -z "$2" ]; then
    echo "No service argument supplied"
    exit 1
fi

if [ -z "$3" ]; then
    echo "No userprovided service argument supplied"
    exit 1
fi

APP_NAME=$1; shift
SERVICE_NAME=$1; shift
USERPROVIDED_SERVICE=$1; shift

app_guid=`cf app $APP_NAME --guid`

tags_formated=`cf curl /v2/apps/$app_guid/env | jq -r '.system_env_json.VCAP_SERVICES."'${SERVICE_NAME}'"[].tags' | sed 's/[][]//g'`

tfile=`mktemp /tmp/credsXXXXXXXXX.json`

cf curl /v2/apps/$app_guid/env | jq -r '.system_env_json.VCAP_SERVICES."'${SERVICE_NAME}'"[].credentials' >> "$tfile"

cf cups $USERPROVIDED_SERVICE -p $tfile -t "$tags_formated"

rm $tfile

