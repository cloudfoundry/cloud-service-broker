#!/usr/bin/env bash

set +x # Hide secrets
set -o nounset
set -o errexit
set -o pipefail

cf push --no-start

APP_NAME=cloud-service-broker

if [[ ${GOOGLE_CREDENTIALS} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_CREDENTIALS "${GOOGLE_CREDENTIALS}"
fi

if [[ ${GOOGLE_PROJECT} ]]; then
  cf set-env "${APP_NAME}" GOOGLE_PROJECT "${GOOGLE_PROJECT}"
fi

if [[ ${ARM_SUBSCRIPTION_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_SUBSCRIPTION_ID "${ARM_SUBSCRIPTION_ID}"
fi

if [[ ${ARM_TENANT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_TENANT_ID "${ARM_TENANT_ID}"
fi

if [[ ${ARM_CLIENT_ID} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_ID "${ARM_CLIENT_ID}"
fi

if [[ ${ARM_CLIENT_SECRET} ]]; then
  cf set-env "${APP_NAME}" ARM_CLIENT_SECRET "${ARM_CLIENT_SECRET}"
fi

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH "${GSB_BROKERPAK_BUILTIN_PATH}"
fi

cf bind-service "${APP_NAME}" csb-sql

cf start "${APP_NAME}"

cf create-service-broker cloud-service-broker "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || echo "broker already registered"
