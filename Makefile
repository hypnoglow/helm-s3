PKG := github.com/hypnoglow/helm-s3
GO111MODULE := on

.EXPORT_ALL_VARIABLES:

.PHONY: all
all: deps build

.PHONY: deps
deps:
	@go mod download
	@go mod vendor
	@go mod tidy

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

.PHONY: test-integration
test-integration:
	@./hack/integration-tests-local.sh

.PHONY: test-e2e
test-e2e:
	go test -v ./tests/e2e/...
