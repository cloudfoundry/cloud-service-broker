set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

SERVICE_NAME=mysql-service-$$

# SERVICE=$1; shift
# PLAN=$1; shift
# #UPDATE_FIELD=$1; shift
# SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$"

SERVICE="csb-aws-mysql"; 
PLAN="small"; 
SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$"

RESULT=1
echo "creating service ..."
cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}"

if wait_for_service "${SERVICE_INSTANCE_NAME}" "create in progress" "create succeeded"; then

   
        jsonfield='{"'instance_name'":"bogus"}'
        # echo $jsonfield
        # Shift all the parameters down by one
        echo "udpating service ..."
        # udpate service..
        
        update_status=$( cf update-service "${SERVICE_INSTANCE_NAME}" -c $jsonfield | grep FAILED)

        echo $update_status

        if [ "$update_status" != "FAILED" ]; then
            delete_service "${SERVICE_INSTANCE_NAME}"
            echo "$0 FAILED"
            exit 1
        fi

        

        delete_service "${SERVICE_INSTANCE_NAME}"
        echo "$0 SUCCEEDED"
        exit 0
    

fi