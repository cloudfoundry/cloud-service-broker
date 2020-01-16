#!/usr/bin/env bash

set +x # Hide secrets

cf push --no-start

APP_NAME=cloud-service-broker

cf set-env "${APP_NAME}" ROOT_SERVICE_ACCOUNT_JSON "${ROOT_SERVICE_ACCOUNT_JSON}"
cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

cf bind-service "${APP_NAME}" csb-sql

cf start "${APP_NAME}"

cf create-service-broker cloud-service-broker "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://cloud-service-broker.cfapps.io --space-scoped
