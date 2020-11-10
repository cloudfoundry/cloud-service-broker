#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

${SCRIPT_DIR}/cf-test-adhoc-services.sh && ${SCRIPT_DIR}/cf-test-spring-music-service.sh
RESULT=$?

if [ ${RESULT} -eq 0 ]; then
    echo "SUCCEEDED: $0"
else
    echo "FAILED: $0"
fi

exit ${RESULT}
