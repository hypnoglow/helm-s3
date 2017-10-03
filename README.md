# helm-s3

[![CircleCI](https://circleci.com/gh/hypnoglow/helm-s3.svg?style=shield)](https://circleci.com/gh/hypnoglow/helm-s3)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

The Helm plugin that provides s3 protocol support. 

This allows you to have private Helm chart repositories hosted on Amazon S3.

## Install

Installation itself is simple as:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git

Plugin requires Golang to be installed to build the binary file from source.
It will happen implicitly on plugin installation, nothing needs to be done manually.

#### Note on AWS authentication

Because this plugin assumes private access to S3, you need to provide valid AWS credentials.
Two options are available:
1) The plugin is able to read AWS default environment variables: `$AWS_ACCESS_KEY_ID`,
`$AWS_SECRET_ACCESS_KEY` and `$AWS_DEFAULT_REGION`.
2) If you already using `aws-cli`, you may already have files `$HOME/.aws/credentials` and `$HOME/.aws/config`.
If so, you are good to go - the plugin can read your credentials from those files.

To minimize security issues, remember to configure your IAM user policies properly - the plugin requires only S3 Read access
on specific bucket.

## Usage

For now let's omit the process of uploading repository index and charts to s3 and assume
you already have your repository `index.yaml` file on s3 under path `s3://bucket-name/charts/index.yaml`
and a chart archive `epicservice-0.5.1.tgz` under path `s3://bucket-name/charts/epicservice-0.5.1.tgz`.

Add your repository:

    $ helm repo add coolcharts s3://bucket-name/charts
    
Now you can use it as any other Helm chart repository.
Try:

    $ helm search coolcharts
    NAME                       	VERSION	  DESCRIPTION
    coolcharts/epicservice	    0.5.1     A Helm chart.
    
    $ helm install coolchart/epicservice --version "0.5.1"

Fetching also works:

    $ helm fetch s3://bucket-name/charts/epicservice-0.5.1.tgz

### Init & Push

To create a new repository, use **init**:

    $ helm s3 init s3://bucket-name/charts

This command generates an empty **index.yaml** and uploads it to the S3 bucket 
under `/charts` key.

To push to this repo by it's name, you need to add it first:

    $ helm repo add mynewrepo s3://bucket-name/charts

Now you can push your chart to this repo:

    $ helm s3 push ./epicservice-0.7.2.tgz mynewrepo

On push, remote repo index is automatically updated. To sync your local index, run:

    $ helm repo update

Now your pushed chart is available:

    $ helm search mynewrepo 
    NAME                    VERSION	 DESCRIPTION
    mynewrepo/epicservice   0.7.2    A Helm chart.

## Uninstall

    $ helm plugin remove s3
    
## Contributing

Contributions are welcome.
    
## License

[MIT](LICENSE)