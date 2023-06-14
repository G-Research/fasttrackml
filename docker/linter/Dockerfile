FROM golangci/golangci-lint

WORKDIR /linter/
ADD . /linter/

CMD ["golangci-lint", "run", "-v"]
