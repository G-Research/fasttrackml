# Build MLFlow UI
FROM --platform=$BUILDPLATFORM node:16 AS mlflow-build

COPY pkg/ui/mlflow /mlflow
RUN /mlflow/build.sh

# Build Aim UI
FROM --platform=$BUILDPLATFORM node:16 AS aim-build

COPY pkg/ui/aim /aim
RUN /aim/build.sh

# Build fasttrack binary
FROM golang:1.20-alpine3.17 AS go-build

ARG tags

RUN apk add --no-cache build-base gcc openssl-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
COPY pkg ./pkg
COPY --from=mlflow-build /mlflow/build ./pkg/ui/mlflow/build
COPY --from=aim-build /aim/build ./pkg/ui/aim/build
RUN go build --tags "$tags"

# Runtime container
FROM alpine:3.17

COPY --from=go-build /build/fasttrack /usr/local/bin/

VOLUME /data
ENV "FASTTRACK_LISTEN_ADDRESS" ":5000"
ENV "FASTTRACK_DATABASE_URI" "sqlite:///data/fasttrack.db"
ENTRYPOINT ["fasttrack"]
CMD ["server"]
