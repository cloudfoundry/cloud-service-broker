#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

allServices=("csb-aws-mysql" "csb-aws-redis" "csb-aws-postgresql")
INSTANCES=()

for s in ${allServices[@]}; do
  create_service "${s}" small "${s}-$$" &
  INSTANCES+=("${s}-$$")
done

NO_TLS_SERVICES=("csb-aws-mysql" "csb-aws-postgresql")

for s in ${NO_TLS_SERVICES[@]}; do
    create_service "${s}" small "${s}-no-tls-$$" "{\"use_tls\":false}" &
    INSTANCES+=("${s}-no-tls-$$")
done

if wait; then
  RESULT=0
  for s in ${INSTANCES[@]}; do
    if [ $# -gt 0 ]; then
      if "${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"; then
        echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-validate-credhub.sh ${s}"
      else
        RESULT=1
        echo "FAILED: ${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"
        break
      fi
    fi
    if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}" medium; then
      echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}"
    else
      RESULT=1
      echo "FAILED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}"
      break
    fi
  done
fi

for s in ${INSTANCES[@]}; do
    delete_service "${s}" &
done

if [ ${RESULT} -eq 0 ]; then
  ${SCRIPT_DIR}/cf-test-s3-bucket.sh
  RESULT=$?
fi

wait

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}






