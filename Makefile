GOLANGCI_LINT_CACHE ?= $(CURDIR)/.cache/golangci-lint
CUSTOM_GOFLAGS ?= -buildvcs=false
E2E_IMAGE ?= go-lint-legibility-e2e
GO_SOURCES = cmd internal plugin tests/e2e

.PHONY: build check custom e2e fmt fmt-check lint lint-golangci test tidy-check vet

check: tidy-check fmt-check vet test e2e lint lint-golangci

build:
	go build -ldflags="-s -w" -trimpath -o bin/go-lint-legibility ./cmd/go-lint-legibility

custom:
	GOFLAGS="$(CUSTOM_GOFLAGS)" golangci-lint custom

e2e:
	docker build --file tests/e2e/Dockerfile --tag "$(E2E_IMAGE)" .
	docker run --rm "$(E2E_IMAGE)"

fmt:
	gofmt -w $(GO_SOURCES)

fmt-check:
	test -z "$$(gofmt -l $(GO_SOURCES))"

lint: build
	./bin/go-lint-legibility ./...

lint-golangci: custom
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) ./bin/legibility-golangci-lint run ./...

test:
	go test ./...

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

vet:
	go vet ./...
