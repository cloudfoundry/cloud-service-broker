#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

ALL_SERVICES=("csb-aws-s3-bucket" "csb-aws-mysql" "csb-aws-redis"ec "csb-aws-postgresql")

# bogus region, node_type should fail
for s in ${ALL_SERVICES[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"region":"bogus"}'
done

INSTANCE_CLASS_SERVICES=("csb-aws-mysql" "csb-aws-postgresql" )

for s in ${INSTANCE_CLASS_SERVICES[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"instance_class":"bogus"}'
done

NODE_TYPE_SERVICES=("csb-aws-redis-basic" "csb-aws-redis")
for s in ${NODE_TYPE_SERVICES[@]}; do
    ../cf-create-service-should-fail.sh ${s} medium '{"node_type":"bogus"}'
done

echo "$0 SUCCEEDED"