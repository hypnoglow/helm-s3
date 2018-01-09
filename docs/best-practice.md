# Best Practice

## Reindexing your repository

In short, due to limitations of AWS your chart repository index can be broken
by accident. This means that it may not reflect the "real" state of your chart 
files in S3 bucket. Nothing serious, but can be annoying.

To workaround this, the `helm s3 reindex <repo>` command is available. *Note: this
operation is is [much more expensive](usage-cost.md#reindex) than other in this plugin*.

## Organizing your repositories

A chart repository file structure is always flat. 
It cannot contain nested directories.

The number of AWS S3 requests for reindex operation depends on your repository structure.
Due to limitations of AWS S3 API you cannot list objects of the folder under the key
 excluding subfolders. `ListObjects` only can lists objects under the key recursively.
 
The plugin code makes its best to ignore subfolders, because chart repository is always flat.
But still, not all cases are covered. 

Imagine the worst case scenario: you have 100 chart files in your repository, which is the
bucket root. And 1 million files in the "foo-bar" subfolder, which are not related to
the chart repository. In this case the plugin **have to** call `ListObjects`
about 1000 times (1000 objects per call) to make sure it did not miss any chart file.

By that, the golden rule is to **never have subfolders in your chart repository folder**.

So, there are two good options for your chart repository file structure inside S3 bucket:

1. One bucket - one repository. Create a bucket "yourcompany-charts-stable", or
"yourcompany-productname-charts" and use the bucket root as your chart repository.
In this case, never put any other files in that bucket.

2. One bucket - many repositories, each in separate subfolder. Create a bucket 
"yourcompany-charts". Create a subfolder in it for each repository you need, for
example "stable" and "testing". Another option is to separate the repositories
by the product or by group of services, for example "backoffice", "order-processing", etc.
And again, never put any other files in the repository folder.