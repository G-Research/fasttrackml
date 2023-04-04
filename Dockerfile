# Build MLFlow UI
FROM --platform=$BUILDPLATFORM node:16 AS mlflow-build

COPY pkg/ui/mlflow/embed /mlflow
RUN /mlflow/build.sh

# Build Aim UI
FROM --platform=$BUILDPLATFORM node:16 AS aim-build

COPY pkg/ui/aim/embed /aim
RUN /aim/build.sh

# Build fasttrack binary
FROM --platform=$BUILDPLATFORM golang:1.20 AS go-build

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
COPY pkg ./pkg
COPY --from=mlflow-build /mlflow/build ./pkg/ui/mlflow/embed/build
COPY --from=aim-build /aim/build ./pkg/ui/aim/embed/build

ARG TARGETARCH
RUN bash -c "\
    GCCARCH=\${TARGETARCH/amd64/x86-64} \
 && GCCARCH=\${GCCARCH/arm64/aarch64} \
 && dpkg --add-architecture $TARGETARCH \
 && apt-get update \
 && apt-get install -y \
    gcc-\$GCCARCH-linux-gnu \
    libc6-dev:$TARGETARCH \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*"

ARG tags=netgo osusergo
ARG version=dev
RUN bash -c "\
    GCCARCH=\${TARGETARCH/amd64/x86_64} \
 && GCCARCH=\${GCCARCH/arm64/aarch64} \
 && CC=\$GCCARCH-linux-gnu-gcc CGO_ENABLED=1 GOARCH=$TARGETARCH \
    go build \
    -tags \"$tags\" \
    -ldflags \"-linkmode external -extldflags '-static' -s -w -X 'github.com/G-Research/fasttrack/pkg/version.Version=$version'\""

# Runtime container
FROM alpine:3.17

COPY --from=go-build /build/fasttrack /usr/local/bin/

VOLUME /data
ENV "FASTTRACK_LISTEN_ADDRESS" ":5000"
ENV "FASTTRACK_DATABASE_URI" "sqlite:///data/fasttrack.db"
ENTRYPOINT ["fasttrack"]
CMD ["server"]
