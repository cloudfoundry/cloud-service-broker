#!/usr/bin/env bash

set -o pipefail
set -o nounset
#set -o errexit

if [ $# -lt 3 ]; then
    echo "Usage: ${0} <resource group> <server name> <location>"
    exit 1
fi

RG=${1}; shift
SERVER_NAME=${1}; shift
LOCATION=${1}; shift

USERNAME=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z' | fold -w 16 | head -n 1)
PASSWORD=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

if [ $# -gt 0 ]; then
    USERNAME=$1; shift
fi

if [ $# -gt 0 ]; then
    PASSWORD=$1; shift
fi

DB_NAME=csb-db

az sql server create --resource-group ${RG} --name ${SERVER_NAME} --location ${LOCATION} --admin-user ${USERNAME} --admin-password ${PASSWORD}

az sql server firewall-rule create --resource-group ${RG} --server ${SERVER_NAME} --name ${SERVER_NAME}-ip --start-ip-address 0.0.0.0 --end-ip-address 0.0.0.0

az sql db create --name ${DB_NAME} --resource-group ${RG} --server ${SERVER_NAME} -e GeneralPurpose -f Gen5 -c 2 --compute-model Serverless --auto-pause-delay 120

DETAILS=$(az sql server show --resource-group ${RG} --name ${SERVER_NAME})

echo Server Details
echo           FQDN: $(echo ${DETAILS} | jq -r .fullyQualifiedDomainName)
echo Admin Username: ${USERNAME}@${SERVER_NAME}
echo Admin Password: ${PASSWORD}
echo  Database Name: ${DB_NAME}
