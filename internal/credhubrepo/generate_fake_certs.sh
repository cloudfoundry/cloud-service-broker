#!/usr/bin/env bash

set -euxo pipefail

# Generate root CA key and cert
openssl req -x509 -newkey rsa:4096 -keyout fakeRootCA.key -out fakeRootCA.crt -sha256 -days 365 -nodes -subj "/C=XX/ST=StateName/O=OrgName/CN=localhost"

# Generate and sign certificate for localhost
openssl genrsa -out fakeLocalhost.key 2048
openssl req -new -sha256 -key fakeLocalhost.key \
  -subj "/C=XX/ST=StateName/O=OrgName/CN=localhost" \
  -reqexts SAN \
  -config <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:localhost")) \
  -out fakeLocalhost.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -in fakeLocalhost.csr -CA fakeRootCA.crt -CAkey fakeRootCA.key -CAcreateserial -out fakeLocalhost.crt -days 365 -sha256