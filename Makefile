pkg := github.com/hypnoglow/helm-s3

build:
	@GOBIN=$(CURDIR)/bin/ go install $(pkg)/cmd/...