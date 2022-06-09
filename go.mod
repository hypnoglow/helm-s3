module github.com/hypnoglow/helm-s3

go 1.16

// See: https://github.com/helm/helm/issues/9354
replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/aws/aws-sdk-go v1.37.18
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.6
	github.com/minio/minio-go/v6 v6.0.40
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	helm.sh/helm/v3 v3.9.0
	k8s.io/helm v2.17.0+incompatible
	sigs.k8s.io/yaml v1.3.0
)

// CVE-2021-25741
require oras.land/oras-go v1.1.1 //indirect

replace (
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.9.2
	oras.land/oras-go => oras.land/oras-go v1.1.1
)
