# helm-s3

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

Let's omit the process of uploading repository index and charts to s3 and assume
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

## Uninstall

    $ helm plugin remove s3
    
## Contributing

Contributions are welcome.
    
## License

[MIT](LICENSE)