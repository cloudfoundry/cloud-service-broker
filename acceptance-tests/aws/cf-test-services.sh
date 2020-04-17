#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

if [ $# -eq 0 ]; then
  echo "Usage: $0 <vpc-id>"
  exit 1
fi

VPC_ID=$1; shift

allServices=("csb-aws-mysql" "csb-aws-redis-basic" "csb-aws-redis-ha")

for s in ${allServices[@]}; do
  create_service "${s}" small "${s}-$$" "{\"aws_vpc_id\": \"${VPC_ID}\"}" &
done

if wait; then
  RESULT=0
  for s in ${allServices[@]}; do
    if [ $# -gt 0 ]; then
      if "${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}-$$"; then
        echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-validate-credhub.sh ${s}-$$"
      else
        RESULT=1
        echo "FAILED: ${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}-$$"
        break
      fi
    fi
    if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"; then
      echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"
    else
      RESULT=1
      echo "FAILED: ${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${s}-$$"
      break
    fi
  done
fi

for s in ${allServices[@]}; do
  delete_service "${s}-$$" &
done

wait

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}






