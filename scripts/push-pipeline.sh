#!/usr/bin/env bash

set -e

lpass sync

STATUS=$(lpass status)
echo "lpass status: ${STATUS}"

if [ -z "${PIPELINE_SECRETS}" ]; then  # Not using concourse - local dev version
    if [[ "${STATUS}" != *Logged?in* ]]; then
      echo -e "\n***WARNING***"
      echo "Not logged in to Last Pass, if you're not prompted for"
      echo "your password you will need to use: "
      echo "  lpass login [--trust] your@email.address"
      exit 1
    fi

    export PIPELINE_SECRETS="`lpass show --notes 'Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml'`"
fi

fly -t hushhouse set-pipeline --pipeline cloud-service-broker --config ci/pipeline.yml -l <(echo "${PIPELINE_SECRETS}")
