module github.com/hypnoglow/helm-s3

go 1.12

// See: https://github.com/helm/helm/issues/6994
replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.0.1
	github.com/aws/aws-sdk-go v1.25.50
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.3.1
	github.com/minio/minio-go/v6 v6.0.40
	github.com/pkg/errors v0.8.1
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/stretchr/testify v1.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.49.0 // indirect
	helm.sh/helm/v3 v3.0.0
	k8s.io/helm v2.16.1+incompatible
	sigs.k8s.io/yaml v1.1.0
)
