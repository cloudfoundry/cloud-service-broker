#!/usr/bin/env bash

set -o errexit
set -o pipefail

if [ -z "$1" ]; then
    echo "No service name argument supplied"
    exit 1
fi

if [ -z "$2" ]; then
    echo "No plane argument supplied"
    exit 1
fi

SERVICE=$1
PLAN=$2

NAME="${SERVICE}-${PLAN}-$$"

set -o nounset

cf create-service "${SERVICE}" "${PLAN}" "${NAME}"

cf service "${NAME}" | grep "create in progress"

set +e
while [ $? -eq 0 ]; do
    sleep 15
    cf service "${NAME}" | grep "create in progress"
done
set -e

APP=spring-music

cf bind-service "${APP}" "${NAME}"

cf restart "${APP}"

cf unbind-service "${APP}" "${NAME}"

cf delete-service -f "${NAME}"

cf service "${NAME}"

set +e
while [ $? -eq 0 ]; do
    sleep 15
    cf service "${NAME}"
done
set -e
