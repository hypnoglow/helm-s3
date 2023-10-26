package main

import (
	"os"
	"time"
)

// options represents global command options (global flags).
type options struct {
	timeout time.Duration
	acl     string
	verbose bool
}

// newDefaultOptions returns default options.
func newDefaultOptions() *options {
	return &options{
		timeout: 5 * time.Minute,
		acl:     os.Getenv("S3_ACL"),
		verbose: false,
	}
}
