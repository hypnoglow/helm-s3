package helmutil

import (
	"io"
)

// Digest hashes a reader and returns a SHA256 digest.
func Digest(in io.Reader) (string, error) {
	if IsHelm3() {
		return digestV3(in)
	}
	return digestV2(in)
}

// DigestFile calculates a SHA256 hash for a given file.
func DigestFile(filename string) (string, error) {
	if IsHelm3() {
		return digestFileV3(filename)
	}
	return digestFileV2(filename)
}
