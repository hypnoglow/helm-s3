package s3object

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		uri              string
		expectedBucket   string
		expectedKey      string
		expectedFilename string
		expectedIDString string
	}{
		{
			uri:            "https://s3.eu-central-1.amazonaws.com/project-omega-bucket-777/testfolder/test1.png",
			expectedBucket: "project-omega-bucket-777",
			expectedKey:    "testfolder/test1.png",
		},
		{
			uri:            "https://project-omega-bucket-777.s3.eu-central-1.amazonaws.com/testfolder/test1.png",
			expectedBucket: "project-omega-bucket-777",
			expectedKey:    "testfolder/test1.png",
		},
	}

	for _, c := range cases {
		bucket, key := Parse(c.uri)

		if c.expectedBucket != bucket {
			t.Errorf("Expected bucket to be %v but got %v", c.expectedBucket, bucket)
		}

		if c.expectedKey != key {
			t.Errorf("Expected key to be %v but got %v", c.expectedKey, key)
		}
	}
}
