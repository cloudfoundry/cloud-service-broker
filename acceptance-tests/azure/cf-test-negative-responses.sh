#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

# bogus location or resource_group should fail
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mysql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-redis medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-redis medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mongodb medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-failover-group medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-eventhubs medium '{"resource_group":"bogus"}'

../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"location":"bogus"}'
../cf-create-service-should-fail.sh csb-azure-mssql-server standard '{"resource_group":"bogus"}'

echo "$0 SUCCEEDED"