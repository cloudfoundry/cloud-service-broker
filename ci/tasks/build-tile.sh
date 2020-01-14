#!/usr/bin/env bash

set -e

export OUTPUT_DIR=$PWD/tiles
export SERVICE_BROKER_DIR=src/cloud-service-broker
export CURRENT_VERSION="$(cat metadata/version)"

apt update
apt install -y zip

mkdir -p tiles

pushd "$SERVICE_BROKER_DIR"
    zip /tmp/cloud-service-broker.zip -r . -x *.git* product/\* release/\* examples/\*
    cp /tmp/cloud-service-broker.zip $OUTPUT_DIR/cloud-service-broker-$CURRENT_VERSION-cf-app.zip

    tile build "$CURRENT_VERSION"
    mv "product/"*.pivotal $OUTPUT_DIR

    tile build "$CURRENT_VERSION-rc"
    mv "product/"*.pivotal $OUTPUT_DIR
popd
