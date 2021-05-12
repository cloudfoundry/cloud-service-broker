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

FROM golang:1.16.4-alpine3.12 AS build
RUN apk update
RUN apk upgrade
RUN apk add --update gcc g++
WORKDIR /app
ADD . /app
RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/cloud-service-broker

FROM alpine:latest

COPY --from=build /app/build/cloud-service-broker /bin/cloud-service-broker

ENV PORT 8080
EXPOSE 8080/tcp

WORKDIR /bin
ENTRYPOINT ["/bin/cloud-service-broker"]
CMD ["help"]
