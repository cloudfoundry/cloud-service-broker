#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

allServices=( "csb-google-mysql" "csb-google-redis" )

# "csb-google-postgres" - does not currently allow second binding

for s in ${allServices[@]}; do
  if [ ${s} == "csb-google-redis" ]
    then
      create_service "${s}" basic "${s}-$$" &
    else
      create_service "${s}" small "${s}-$$" &
  fi
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

if [ ${RESULT} -eq 0 ]; then
  ${SCRIPT_DIR}/cf-test-stack-driver.sh && ${SCRIPT_DIR}/cf-test-dataproc.sh
  RESULT=$?
fi

wait

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}
