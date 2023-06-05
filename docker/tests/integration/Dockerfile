FROM golang:1.20

WORKDIR /go/src
COPY . /go/src

RUN make go-get

CMD ["make", "test-go-integration"]
