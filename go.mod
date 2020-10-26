module github.com/hypnoglow/helm-s3

go 1.15

// See: https://github.com/helm/helm/issues/6994
replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/aws/aws-sdk-go v1.27.0
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.4.0
	github.com/minio/minio-go/v6 v6.0.40
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.49.0 // indirect
	helm.sh/helm/v3 v3.4.0
	k8s.io/helm v2.17.0+incompatible
	sigs.k8s.io/yaml v1.2.0
)
