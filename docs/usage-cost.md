# Usage pricing

I hope this document helps you to calculate the AWS S3 usage cost for your use case. 

Disclaimer: the plugin author is not responsible for your unexpected expenses.

**Make sure to consult the pricing for your region [here](https://aws.amazon.com/s3/pricing)!** 

## Reindex

`helm s3 reindex <repo>` command is much more expensive operation than other in
this plugin. For example, reindexing a repository with 1000 chart files in it 
results in 1 GET (`ListObjects`) request and 1000 HEAD (`HeadObject`) requests.
Plus it can make additional GET (`GetObject`) requests if it did not found 
required metadata in the HEAD request response.

At the moment of writing this document the price for HEAD/GET requests in `eu-central-1` is `$0.0043 for 10 000 requests`.
So the whole reindex operation for this case may cost approximately **$0.00043** or even **$0.00086**. 
This seems small, but multiple reindex operations per day may hurt your budget. 