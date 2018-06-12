PKG := github.com/hypnoglow/helm-s3

.PHONY: dep
dep:
	@dep ensure -v -vendor-only

.PHONY: build
build:
	@./hack/build.sh $(CURDIR) $(PKG)

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
