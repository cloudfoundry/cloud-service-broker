#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

if [ $# -eq 0 ]; then
  echo "Usage: $0 <network name>"
  exit 1
fi

NETWORK=$1; shift

allServices=(  "csb-google-mysql" "csb-google-redis" "csb-google-postgres")

for s in ${allServices[@]}; do
  
  if [  ${s} == "csb-google-redis"  ]
    then
        create_service "${s}" basic "${s}-$$" "{\"authorized_network\": \"${NETWORK}\"}" &
    else
       create_service "${s}" small "${s}-$$" "{\"authorized_network\": \"${NETWORK}\"}" &
  fi

done

if wait; then
  RESULT=0
  for s in ${allServices[@]}; do
      if [  ${s} != "csb-google-redis"  ] || [  ${s} != "csb-google-postgres" ]; then    
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
