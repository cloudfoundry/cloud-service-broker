#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

allServices=("csb-aws-mysql" )

# bogus location or resource_group should fail
for s in ${allServices[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"region":"bogus"}'
done

echo "$0 SUCCEEDED"