package main

import "github.com/hypnoglow/helm-s3/internal/awsutil"

type printer interface {
	Printf(format string, v ...interface{})
	PrintErrf(format string, i ...interface{})
}

// escapeIfRelative escapes chart filename if it is indexed as relative.
//
// Note: we escape filename only if 'relative' is set for a few reasons:
//   - Full URLs don't need to be escaped because with full URLs in the index
//     the charts can be downloaded only with this plugin;
//   - Even if we escape the filename here, the code in Index.Add and
//     Index.AddOrReplace (in particular, the call to urlutil.URLJoin) will
//     break the URL, e.g. the escaped filename "petstore-1.0.0%2B102.tgz"
//     with the "s3://example-bucket" baseURL will become
//     "s3://example-bucket/petstore-1.0.0%252B102.tgz".
//     So if we ever decide to escape, we need to fix this.
func escapeIfRelative(filename string, relative bool) string {
	if !relative {
		return filename
	}

	return awsutil.EscapePath(filename)
}
