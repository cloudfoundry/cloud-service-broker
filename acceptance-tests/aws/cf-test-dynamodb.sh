#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"



SERVICE_NAME=dynamodb-$$



RESULT=1
if create_service csb-aws-dynamodb ondemand "${SERVICE_NAME}" ${SCRIPT_DIR}/config.json; then

    (cd "${SCRIPT_DIR}/springboot-dynamodb" && cf push --no-start)
    if cf bind-service dynamodb-demo "${SERVICE_NAME}"; then
        if cf start dynamodb-demo; then
            RESULT=0
            echo "springboot-dynamodb success"
            response=$(curl --write-out %{http_code} --silent --output /dev/null $(cf app dynamodb-demo | grep 'routes:' | cut -d ':' -f 2 | xargs)"/testdynamodb")
            echo $response
            if [ "$response" = "200" ]
            then
            echo "Dynamodb success"
            else 
                RESULT=1
                echo "springboot-dynamodb failed to access dynamodb "
            fi
        else
            echo "springboot-dynamodb failed"
            cf logs dynamodb-demo --recent
        fi
        #cf delete -f javagcpapp-demo 
    fi
    #delete_service ${SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi



cf delete -f dynamodb-demo 
delete_service ${SERVICE_NAME}

exit ${RESULT}