#!/usr/bin/env bash

set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

PRIMARY_SERVER_NAME=mssql-server-p-$$
SECONDARY_SERVER_NAME=mssql-server-s-$$
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~

SERVER_RG=rg-test-service-$$
RESULT=1
if create_service csb-azure-resource-group standard "${SERVER_RG}" "{\"instance_name\":\"${SERVER_RG}\"}"; then
    "${SCRIPT_DIR}/cf-create-mssql-server.sh" "${PRIMARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${SERVER_RG}" eastus &
    PRIMARY_PID=$!

    "${SCRIPT_DIR}/cf-create-mssql-server.sh" "${SECONDARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${SERVER_RG}" westus &
    SECONDARY_PID=$!

    if wait ${PRIMARY_PID} && wait ${SECONDARY_PID}; then
        FOG_NAME=mssql-server-fog-$$
        CONFIG="{
            \"instance_name\":\"${FOG_NAME}\", \
            \"db_name\":\"test_db\", \
            \"server_pair\":\"test\" \
        }"

        MSSQL_DB_FOG_SERVER_PAIR_CREDS="{ \
            \"test\":{ \
                \"admin_username\":\"${USERNAME}\", \
                \"admin_password\":\"${PASSWORD}\", \
                \"primary\":{ \
                  \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                  \"resource_group\":\"${SERVER_RG}\" \
                }, \
                \"secondary\":{ \
                  \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                  \"resource_group\":\"${SERVER_RG}\" \
                } \
            } \
        }" 

        GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS="{ \
            \"server_credential_pairs\":${MSSQL_DB_FOG_SERVER_PAIR_CREDS} \
        }"

        cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS "${GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS}"
        cf set-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS "${MSSQL_DB_FOG_SERVER_PAIR_CREDS}"
        cf restart cloud-service-broker        

        FOG_INSTANCE=test-fog-$$
        FOG_DR_INSTANCE=test-fog-dr-$$
        if create_service csb-azure-mssql-db-failover-group medium "${FOG_INSTANCE}" "${CONFIG}"; then
            if ${SCRIPT_DIR}/../cf-run-spring-music-test.sh "${FOG_INSTANCE}" large "${CONFIG}"; then
                if create_service csb-azure-mssql-db-failover-group existing "${FOG_DR_INSTANCE}" "${CONFIG}"; then
                    if ${SCRIPT_DIR}/../cf-run-spring-music-test.sh "${FOG_DR_INSTANCE}"; then
                        echo Test bind/unbind after changing admin passwords
                        NEW_ADMIN_PASSWORD=Another_S0uld-3eC0mplex~
                        update_service_params "${PRIMARY_SERVER_NAME}" "{\"admin_password\":\"${NEW_ADMIN_PASSWORD}\"}"
                        update_service_params "${SECONDARY_SERVER_NAME}" "{\"admin_password\":\"${NEW_ADMIN_PASSWORD}\"}"

                        MSSQL_DB_FOG_SERVER_PAIR_CREDS="{ \
                            \"test\":{ \
                                \"admin_username\":\"${USERNAME}\", \
                                \"admin_password\":\"${NEW_ADMIN_PASSWORD}\", \
                                \"primary\":{ \
                                  \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                                  \"resource_group\":\"${SERVER_RG}\" \
                                }, \
                                \"secondary\":{ \
                                  \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                                  \"resource_group\":\"${SERVER_RG}\" \
                                } \
                            } \
                        }"         

                        GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS="{ \
                            \"server_credential_pairs\":${MSSQL_DB_FOG_SERVER_PAIR_CREDS} \
                        }"                     
                        cf set-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS "${GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS}"
                        cf set-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS "${MSSQL_DB_FOG_SERVER_PAIR_CREDS}"
                        cf restart cloud-service-broker         
                        if ${SCRIPT_DIR}/../cf-run-spring-music-test.sh "${FOG_DR_INSTANCE}"; then
                            RESULT=0
                        else
                            echo "spring music test failed after admin password change"
                        fi  
                    else
                        echo "spring music test failed on FOG DR"
                    fi
                    delete_service "${FOG_DR_INSTANCE}"
                else
                    echo "failed to create FOG DR"
                fi
            else
                echo "spring music test failed on FOG"
            fi
            delete_service "${FOG_INSTANCE}"
        else
            echo "failed to create FOG"
        fi

        cf unset-env cloud-service-broker GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS
        cf unset-env cloud-service-broker MSSQL_DB_FOG_SERVER_PAIR_CREDS
        cf restart cloud-service-broker           
    fi

    "${SCRIPT_DIR}/cf-delete-mssql-server.sh" "${PRIMARY_SERVER_NAME}" &
    "${SCRIPT_DIR}/cf-delete-mssql-server.sh" "${SECONDARY_SERVER_NAME}" &

    wait

    delete_service "${SERVER_RG}"
else
    echo "Failed creating resource group ${SERVER_RG} for test services"
fi

exit ${RESULT}
