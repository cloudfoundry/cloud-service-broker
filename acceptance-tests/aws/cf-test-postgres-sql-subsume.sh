#!/usr/bin/env bash

set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

echo $#

if [ $# -lt 3 ]; then
  echo "usage: $0 <server name> <region> <admin-password>"
  exit 1
fi

SERVER_NAME=$1
REGION=$2
SERVER_ADMIN_PASSWORD=$3
DB_NAME=subsume-test-db-$$
SUBSUMED_INSTANCE_NAME=csb-subsume-$$
STORAGE=22;


CSB_INSTANCE_NAME=csb-db-$$
CSB_DB_CONFIG="{ \
  \"aws_db_id\": \"${SERVER_NAME}\", \
  \"region\": \"${REGION}\", \
  \"admin_password\": \"${SERVER_ADMIN_PASSWORD}\" \
}"

UPDATE_CONFIG="{ \
  \"storage_gb\": "${STORAGE}", \
  \"region\": \"${REGION}\" \
}"  

echo $UPDATE_CONFIG
RESULT=1


if create_service csb-aws-postgresql subsume "${SUBSUMED_INSTANCE_NAME}" "${CSB_DB_CONFIG}"; then

    (cd "${SCRIPT_DIR}/importtestapp" && cf push --no-start)
    if cf bind-service javaawsapp-demo "${SUBSUMED_INSTANCE_NAME}" ; then
        if cf start javaawsapp-demo; then
            RESULT=0
            echo "javaawsapp-demo success"
            response=$(curl --write-out %{http_code} --silent --output /dev/null $(cf app javaawsapp-demo | grep 'routes:' | cut -d ':' -f 2 | xargs)"/testpostgres")
            echo $response
            if [ "$response" = "200" ]
            then
            echo "javaawsapp-demo success"
            else 
                RESULT=1
                echo "javaawsapp-demo failed to access postgres"
            fi
        else
            echo "javaawsapp-demo failed"
            cf logs javaawsapp-demo --recent
        fi
        cf delete -f javaawsapp-demo
        cf purge-service-instance "${SUBSUMED_INSTANCE_NAME}" -f
    fi
   
fi

# for update we need to increment storage +1 each time we update or AWS won't let update the storage. 

# if update_service_params "${SUBSUMED_INSTANCE_NAME}" "${UPDATE_CONFIG}"; then
#     echo "subsumed postgres instance update successful"  
    
#     cf purge-service-instance "${SUBSUMED_INSTANCE_NAME}" -f
# else 
#      echo "failed to update subsumed postgres instance"
# fi


if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi


exit 0;