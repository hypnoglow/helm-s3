PKG := github.com/hypnoglow/helm-s3

.PHONY: dep
dep:
	@dep ensure -v -vendor-only

.PHONY: build
build:
	@./sh/build.sh $(CURDIR) $(PKG)

.PHONY: install
install:
	@./sh/install.sh

.PHONY: test
test:
	go test ./...

.PHONY: test-integration
test-integration:
	@./sh/integration-tests-local.sh
