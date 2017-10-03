pkg := github.com/hypnoglow/helm-s3

build:
	@./sh/build.sh $(CURDIR) $(pkg)

install:
	@./sh/install.sh