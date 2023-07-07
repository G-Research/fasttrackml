# Build fml binary
FROM --platform=$BUILDPLATFORM golang:1.20 AS go-build

WORKDIR /build
COPY go.mod go.sum /build
RUN go mod download
COPY main.go /build
COPY pkg /build

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
    -o fml \
    -tags \"$tags\" \
    -ldflags \"-linkmode external -extldflags '-static' -s -w -X 'github.com/G-Research/fasttrackml/pkg/version.Version=$version'\""

# Runtime container
FROM alpine:3.18

COPY --from=go-build /build/fml /usr/local/bin/

VOLUME /data
ENV "FML_LISTEN_ADDRESS" ":5000"
ENV "FML_DATABASE_URI" "sqlite:///data/fasttrackml.db"
ENTRYPOINT ["fml"]
CMD ["server"]
