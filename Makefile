PKG := github.com/hypnoglow/helm-s3

.PHONY: all
all: deps build

.PHONY: deps
deps:
	@go mod tidy
	@go mod vendor

.PHONY: build
build:
	@./hack/build.sh $(CURDIR) $(PKG)

.PHONY: build-local
build-local:
	HELM_S3_PLUGIN_VERSION=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") $(MAKE) build

.PHONY: install
install:
	@./hack/install.sh

.PHONY: test-unit
test-unit:
	go test $$(go list ./... | grep -v e2e)

.PHONY: test-e2e
test-e2e:
	go test -v ./tests/e2e/...

.PHONY: test-e2e-local
test-e2e-local:
	@./hack/test-e2e-local.sh

.PHONY: lint
lint:
	@golangci-lint run
