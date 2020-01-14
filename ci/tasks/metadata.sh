#!/bin/sh

set -e

echo "Installing dependencies"
apk update
apk add git

echo "Generating metadata"
mkdir -p metadata/docs

git --git-dir=cloud-service-broker/.git rev-parse HEAD > metadata/revision
./compiled-broker/cloud-service-broker version > metadata/version
./compiled-broker/cloud-service-broker generate tile > metadata/tile.yml
./compiled-broker/cloud-service-broker generate use > metadata/manifest.yml
./compiled-broker/cloud-service-broker generate customization > metadata/docs/customization.md
./compiled-broker/cloud-service-broker generate use --destination-dir="metadata/docs/"
