package s3object

import (
	"net/url"
	"strings"
)

func Parse(uri string) (bucket, key string) {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	// Image URI can be in two forms:
	// https://s3.eu-central-1.amazonaws.com/project-omega-bucket-777/testfolder/test1.png
	// https://project-omega-bucket-777.s3.eu-central-1.amazonaws.com/testfolder/test1.png

	if strings.HasPrefix(u.Host, "s3.") {
		parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
		bucket = parts[0]
		key = parts[1]
	} else {
		parts := strings.SplitN(u.Host, ".", 2)
		bucket = parts[0]
		key = strings.TrimPrefix(u.Path, "/")
	}

	return bucket, key
}
