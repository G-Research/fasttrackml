FROM node:16 AS mlflow-build

COPY ui/mlflow /mlflow
RUN /mlflow/build.sh

FROM golang:1.19-alpine3.17 AS go-build

RUN apk add build-base gcc openssl-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=mlflow-build /mlflow/build ./ui/mlflow/build
RUN go build --tags "sqlcipher sqlite_unlock_notify sqlite_foreign_keys sqlite_vacuum_incr"


FROM alpine:3.17

COPY --from=go-build /build/fasttrack /usr/local/bin/

VOLUME /data
ENV "FASTTRACK_LISTEN_ADDRESS" ":5000"
ENV "FASTTRACK_DATABASE_URI" "sqlite:///data/fasttrack.db"
ENTRYPOINT ["fasttrack"]
CMD ["server"]
