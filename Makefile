pkg := github.com/hypnoglow/helm-s3

dep:
	@dep ensure -v -vendor-only

build:
	@./sh/build.sh $(CURDIR) $(pkg)

install:
	@./sh/install.sh

integration-tests:
	@./sh/integration-tests-local.sh
