#!/usr/bin/env bash

set -o nounset

NAME=$1; shift

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

delete_service "${NAME}"

exit $?