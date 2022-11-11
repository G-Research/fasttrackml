FROM golang:1.19.2 AS build

RUN apt-get update && apt-get install -y libssl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build --tags "sqlcipher sqlite_unlock_notify sqlite_foreign_keys sqlite_vacuum_incr"

FROM ubuntu:20.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y libssl1.1 && rm -rf /var/lib/apt/lists/*

COPY --from=build /build/fasttrack /usr/local/bin/

ENTRYPOINT ["fasttrack"]
