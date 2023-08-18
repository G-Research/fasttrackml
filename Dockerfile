# Build fml binary
FROM --platform=$BUILDPLATFORM golang:1.21 AS go-build

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

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY Makefile .go-build-tags main.go .
COPY pkg ./pkg

ARG version=dev
RUN bash -c "\
    GCCARCH=\${TARGETARCH/amd64/x86_64} \
 && GCCARCH=\${GCCARCH/arm64/aarch64} \
 && CC=\$GCCARCH-linux-gnu-gcc GOARCH=$TARGETARCH VERSION=$version \
    make build"

# Runtime container
FROM alpine:3.18

COPY --from=go-build /build/fml /usr/local/bin/

VOLUME /data
ENV "FML_LISTEN_ADDRESS" ":5000"
ENV "FML_DATABASE_URI" "sqlite:///data/fasttrackml.db"
ENTRYPOINT ["fml"]
CMD ["server"]
