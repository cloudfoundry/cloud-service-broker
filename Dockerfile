# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:alpine AS build
# Golang Docker images come with `GOTOOLCHAIN=local` by default which prevents downloading toolchains on-the-fly.
# This would cause our CI to fail for a few hours when a new go version is released but golang tags are not yet up-to-date with it.
# https://github.com/search?q=repo%3Adocker-library%2Fgolang%20GOTOOLCHAIN&type=code
ENV GOTOOLCHAIN=
RUN apk update
RUN apk upgrade
RUN apk add --update gcc g++
WORKDIR /app
ADD . /app

ARG CSB_VERSION=0.0.0
RUN CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CGO_ENABLED=1 GOOS=linux go build -o ./build/cloud-service-broker -ldflags "-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$CSB_VERSION"

FROM alpine:3.22

COPY --from=build /app/build/cloud-service-broker /bin/cloud-service-broker

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
ADD https://s3.amazonaws.com/rds-downloads/rds-ca-2019-root.pem /usr/local/share/ca-certificates/
RUN update-ca-certificates

ENV PORT 8080
EXPOSE 8080/tcp

WORKDIR /bin
ENTRYPOINT ["/bin/cloud-service-broker"]
CMD ["help"]
