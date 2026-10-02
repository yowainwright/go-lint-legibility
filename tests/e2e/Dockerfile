# syntax=docker/dockerfile:1

FROM golang:1.27.1-bookworm

WORKDIR /workspace

COPY .custom-gcl.yml go.mod go.sum ./

RUN version="$(sed -n 's/^version: //p' .custom-gcl.yml)" \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"${version}" \
    && go mod download

COPY . .

RUN GOFLAGS=-buildvcs=false golangci-lint custom

ENV LEGIBILITY_E2E_BINARY=/workspace/bin/legibility-golangci-lint

CMD ["go", "test", "-count=1", "-v", "-tags=e2e", "./tests/e2e"]
