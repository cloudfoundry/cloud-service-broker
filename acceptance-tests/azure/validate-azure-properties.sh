#!/usr/bin/env bash

set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export DOCKER_COMMON="--rm -u $UID -v $HOME:$HOME -e HOME -e USER=$USER -e USERNAME=$USER -i"
alias cf='docker run $DOCKER_COMMON -w $PWD -t cfplatformeng/cf:latest'
alias az='docker run $DOCKER_COMMON -w $PWD -t microsoft/azure-cli'


