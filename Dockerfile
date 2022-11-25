FROM node:16 AS js-build

WORKDIR /js
COPY js/yarn ./yarn
COPY js/package.json js/yarn.lock js/.yarnrc.yml ./
RUN yarn install
COPY js ./
RUN yarn build


FROM golang:1.19.2 AS go-build

RUN apt-get update && apt-get install -y libssl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=js-build /js/build ./js/build
RUN go build --tags "sqlcipher sqlite_unlock_notify sqlite_foreign_keys sqlite_vacuum_incr"


FROM ubuntu:20.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y libssl1.1 && rm -rf /var/lib/apt/lists/*

COPY --from=go-build /build/fasttrack /usr/local/bin/

ENTRYPOINT ["fasttrack"]
