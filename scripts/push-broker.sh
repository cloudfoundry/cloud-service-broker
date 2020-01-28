#!/usr/bin/env bash

set +x # Hide secrets
set -o nounset
set -o errexit
set -o pipefail

cf push --no-start

APP_NAME=cloud-service-broker

cf set-env "${APP_NAME}" GOOGLE_CREDENTIALS "${GOOGLE_CREDENTIALS}"
cf set-env "${APP_NAME}" GOOGLE_PROJECT "${GOOGLE_PROJECT}"
cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

cf bind-service "${APP_NAME}" csb-sql

cf start "${APP_NAME}"

cf create-service-broker cloud-service-broker "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || echo "broker already registered"
