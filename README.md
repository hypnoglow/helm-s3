# helm-s3

The Helm plugin that provides s3 protocol support. 

This allows you to have private Helm chart repositories hosted on Amazon S3.

## Install

Installation itself is simple as:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git

Plugin requires Golang to be installed to build the binary file from source.
It will happen implicitly on plugin installation, nothing needs to be done manually.

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
    coolcharts/epicservice	        0.5.1     A Helm chart.
    
    $ helm install coolchart/epicservice --version "0.5.1"

Fetching also works:

    $ helm fetch s3://bucket-name/charts/epicservice-0.5.1.tgz

## Uninstall

    $ helm plugin remove s3
    
## Contributing

Contributions are welcome.
    
## License

[MIT](LICENSE)