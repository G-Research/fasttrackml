#
# Build arguments.
#
ARG GOSUMDB=off
ARG PATH_GO_SOURCES=/go/src

#
# Build Go binary inside base container.
#
FROM golang:1.20
# Stage arguments
ARG PATH_GO_SOURCES
# Env variables.
ENV GOSUMDB=$GOSUMDB

# Create sources directory inside the container and copy project files.
RUN mkdir -p $PATH_GO_SOURCES/
WORKDIR $PATH_GO_SOURCES
COPY . $PATH_GO_SOURCES

# Build
RUN make go-get
RUN make go-build

# Container settings.
ENV LISTEN-ADDRESS 8080
EXPOSE 8080
ENTRYPOINT ["./fasttrack-service", "server", "--database-uri=postgres://postgres:postgres@fasttrack-postgres:5432/postgres", "--listen-address=:8080"]
