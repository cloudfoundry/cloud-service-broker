#!/usr/bin/env bash

set -euo pipefail

function main {
  testOutput "Building upgrade-all-service-instances plugin"

  SCRIPTS_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

  BUILD_DIR=`mktemp -d`

  pushd "$SCRIPTS_DIR/.."
      go get -u ./...
      go build -o "$BUILD_DIR/upgrade-all-service-instance-plugin-dev"
  popd

  testOutput "Installing upgrade-all-service-instances plugin"
  cf install-plugin "$BUILD_DIR/upgrade-all-service-instance-plugin-dev" -f

  testOutput "Checking upgrade-all-service-instances plugin is usable"
  cf upgrade-all-service-instances --help

  testOutput "Uninstalling upgrade-all-service-instances plugin"
  cf uninstall-plugin UpgradeAllServiceInstances

  testOutput "Test Success"
}

function testOutput {
  echo -e "\n\n--------\n"$1"\n--------\n\n"
}

main