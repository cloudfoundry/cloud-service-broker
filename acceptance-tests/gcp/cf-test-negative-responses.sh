#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

ALL_SERVICES=("google-redis" "google-mysql")

# bogus region
for s in ${ALL_SERVICES[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"region":"bogus"}'
done

echo "$0 SUCCEEDED"