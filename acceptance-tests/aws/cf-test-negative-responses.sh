#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

allServices=("csb-aws-mysql" "csb-aws-redis" )

# bogus region, node_type should fail
for s in ${allServices[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"region":"bogus"}'
done

../cf-create-service-should-fail.sh csb-aws-mysql medium '{"instance_class":"bogus"}'

../cf-create-service-should-fail.sh csb-aws-redis medium '{"node_type":"bogus"}'

echo "$0 SUCCEEDED"