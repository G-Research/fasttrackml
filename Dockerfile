FROM node:16 AS js-build

WORKDIR /js
COPY js/yarn ./yarn
COPY js/vendor ./vendor
COPY js/package.json js/yarn.lock js/.yarnrc.yml ./
RUN yarn install
COPY js ./
RUN yarn build


FROM golang:1.19-bullseye AS go-build

RUN apt-get update && apt-get install -y libssl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=js-build /js/build ./js/build
RUN CGO_LDFLAGS="/usr/lib/$(uname -m)-linux-gnu/libcrypto.a" go build --tags "sqlcipher sqlite_unlock_notify sqlite_foreign_keys sqlite_vacuum_incr"


FROM debian:bullseye

COPY --from=go-build /build/fasttrack /usr/local/bin/

ENTRYPOINT ["fasttrack"]
