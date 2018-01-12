package awsutil

import (
	"fmt"
	"os"
)

// StderrTokenProvider implements token provider for AWS SDK.
func StderrTokenProvider() (string, error) {
	var v string
	fmt.Fprintf(os.Stderr, "Assume Role MFA token code: ")
	_, err := fmt.Fscanln(os.Stderr, &v)

	return v, err
}
