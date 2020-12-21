#!/usr/bin/env bash

set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

SERVER_NAME=mssql-server-$$
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~

SERVER_RG=rg-test-service-$$
RESULT=1
if create_service csb-azure-resource-group standard "${SERVER_RG}" "{\"instance_name\":\"${SERVER_RG}\"}"; then
    if "${SCRIPT_DIR}/cf-create-mssql-server.sh" "${SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${SERVER_RG}" centralus; then
        CONFIG="{ \
            \"server\": \"test_server\" \
          }"
        
        MSSQL_DB_SERVER_CREDS="{ \
                  \"test_server\": { \
                    \"server_name\":\"${SERVER_NAME}\", \
                    \"admin_username\":\"${USERNAME}\", \
                    \"admin_password\":\"${PASSWORD}\", \
                    \"server_resource_group\":\"${SERVER_RG}\" \
                  }, \
                  \"fail_server\": { \
                    \"server_name\":\"missing\", \
                    \"admin_username\":\"bogus\", \
                    \"admin_password\":\"bad-password\", \
                    \"server_resource_group\":\"rg\" \
                  } \
                }"  

        GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS="{ \"server_credentials\": ${MSSQL_DB_SERVER_CREDS} }"

        echo $CONFIG
        echo $GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS

        cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS "${GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS}"
        cf set-env cloud-service-broker MSSQL_DB_SERVER_CREDS "${MSSQL_DB_SERVER_CREDS}"
        cf restart cloud-service-broker

        MSSQL_DB_INSTANCE=mssql-db-$$
        create_service csb-azure-mssql-db small ${MSSQL_DB_INSTANCE} "${CONFIG}"

        echo Test basic functionality...
        if ${SCRIPT_DIR}/../cf-run-spring-music-test.sh "${MSSQL_DB_INSTANCE}"; then

            NEW_ADMIN_PASSWORD=Another_S0uld-3eC0mplex~
            update_service_params "${SERVER_NAME}" "{\"admin_password\":\"${NEW_ADMIN_PASSWORD}\"}"

            MSSQL_DB_SERVER_CREDS="{ \
                  \"test_server\": { \
                    \"server_name\":\"${SERVER_NAME}\", \
                    \"admin_username\":\"${USERNAME}\", \
                    \"admin_password\":\"${NEW_ADMIN_PASSWORD}\", \
                    \"server_resource_group\":\"${SERVER_RG}\" \
                  }, \
                  \"fail_server\": { \
                    \"server_name\":\"missing\", \
                    \"admin_username\":\"bogus\", \
                    \"admin_password\":\"bad-password\", \
                    \"server_resource_group\":\"rg\" \
                  } \
                }"
            NEW_GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS="{ \"server_credentials\": ${MSSQL_DB_SERVER_CREDS} }"

            cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS "${NEW_GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS}"
            cf set-env cloud-service-broker MSSQL_DB_SERVER_CREDS "${MSSQL_DB_SERVER_CREDS}"
            cf restart cloud-service-broker

            echo Test bind/unbind after admin password change...
            if ${SCRIPT_DIR}/../cf-run-spring-music-test.sh "${MSSQL_DB_INSTANCE}"; then
                echo Test unbind works after changing admin password
                if cf bind-service spring-music ${MSSQL_DB_INSTANCE}; then
                    NEW_ADMIN_PASSWORD=YetAnother_S0uld-3eC0mplex~
                    update_service_params "${SERVER_NAME}" "{\"admin_password\":\"${NEW_ADMIN_PASSWORD}\"}"

                    MSSQL_DB_SERVER_CREDS="{ \
                          \"test_server\": { \
                            \"server_name\":\"${SERVER_NAME}\", \
                            \"admin_username\":\"${USERNAME}\", \
                            \"admin_password\":\"${NEW_ADMIN_PASSWORD}\", \
                            \"server_resource_group\":\"${SERVER_RG}\" \
                          }, \
                          \"fail_server\": { \
                            \"server_name\":\"missing\", \
                            \"admin_username\":\"bogus\", \
                            \"admin_password\":\"bad-password\", \
                            \"server_resource_group\":\"rg\" \
                          } \
                        }"
                    NEW_GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS="{ \"server_credentials\": ${MSSQL_DB_SERVER_CREDS} }"

                    cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS "${NEW_GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS}"
                    cf set-env cloud-service-broker MSSQL_DB_SERVER_CREDS "${MSSQL_DB_SERVER_CREDS}"
                    cf restart cloud-service-broker

                    if cf unbind-service spring-music ${MSSQL_DB_INSTANCE}; then
                        RESULT=$?
                    else
                        echo Unbind after password change failed!
                    fi
                fi
            else
                echo Bind/unbind failed after password change!
            fi
            cf unset-env cloud-service-broker MSSQL_DB_SERVER_CREDS
        else
            echo Basic functionality failed!
        fi
        delete_service "${MSSQL_DB_INSTANCE}"

        cf unset-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS
        cf restart cloud-service-broker
    fi

    "${SCRIPT_DIR}/cf-delete-mssql-server.sh" "${SERVER_NAME}"

    delete_service "${SERVER_RG}"
else
    echo "Failed creating resource group ${SERVER_RG} for test services"
fi

exit ${RESULT}
