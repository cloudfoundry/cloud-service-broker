#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"



SERVICE_NAME=storage-bucket-$$
BIG_QUERY_SERVICE_NAME=my-big-query-$$
SPANNER_SERVICE_NAME=my-spanner-$$

CSB_BUCKET_CONFIG="{ \
  \"role\": \"storage.objectAdmin\" \
}"

RESULT=1
if create_service csb-google-storage-bucket private "${SERVICE_NAME}" "${CSB_BUCKET_CONFIG}"; then

    (cd "${SCRIPT_DIR}/java-gcp-apps" && cf push --no-start)
    if cf bind-service javagcpapp-demo "${SERVICE_NAME}" -c "${CSB_BUCKET_CONFIG}"; then
        if cf start javagcpapp-demo; then
            RESULT=0
            echo "javagcpapp-demo success"
            response=$(curl --write-out %{http_code} --silent --output /dev/null $(cf app javagcpapp-demo | grep 'routes:' | cut -d ':' -f 2 | xargs)"/testgcpbucket")
            echo $response
            if [ "$response" = "200" ]
            then
            echo "javagcpapp-demo success"
            else 
                RESULT=1
                echo "javagcpapp-demo failed to access bucket"
            fi
        else
            echo "javagcpapp-demo failed"
            cf logs javagcpapp-demo --recent
        fi
        cf delete -f javagcpapp-demo 
    fi
    delete_service ${SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi




RESULT=1
if create_service csb-google-bigquery standard "${BIG_QUERY_SERVICE_NAME}" ; then

    (cd "${SCRIPT_DIR}/java-gcp-apps" && cf push --no-start)
    if cf bind-service javagcpapp-demo "${BIG_QUERY_SERVICE_NAME}" ; then
        if cf start javagcpapp-demo; then
            RESULT=0
            echo "javagcpapp-demo success"
            response=$(curl --write-out %{http_code} --silent --output /dev/null $(cf app javagcpapp-demo | grep 'routes:' | cut -d ':' -f 2 | xargs)"/testgcpbigquery")
            echo $response
            if [ "$response" = "200" ]
            then
            echo "javagcpapp-demo success"
            else 
                RESULT=1
                echo "javagcpapp-demo failed to access bigquery"
            fi
        else
            echo "javagcpapp-demo failed"
            cf logs javagcpapp-demo --recent
        fi
        cf delete -f javagcpapp-demo 
    fi
    delete_service ${BIG_QUERY_SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi


RESULT=1
if create_service csb-google-spanner small "${SPANNER_SERVICE_NAME}" ; then

    (cd "${SCRIPT_DIR}/java-gcp-apps" && cf push --no-start)
    if cf bind-service javagcpapp-demo "${SPANNER_SERVICE_NAME}" ; then
        if cf start javagcpapp-demo; then
            RESULT=0
            echo "javagcpapp-demo success"
            response=$(curl --write-out %{http_code} --silent --output /dev/null $(cf app javagcpapp-demo | grep 'routes:' | cut -d ':' -f 2 | xargs)"/testgcpspanner")
            echo $response
            if [ "$response" = "200" ]
            then
            echo "javagcpapp-demo success"
            sleep 30
            else 
                RESULT=1
                echo "javagcpapp-demo failed to access gcpspanner"

            fi
        else
            echo "javagcpapp-demo failed"
            cf logs javagcpapp-demo --recent
        fi
        cf delete -f javagcpapp-demo 
    fi
    delete_service ${SPANNER_SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi

exit ${RESULT}