package b2_test

import (
	"fmt"
	. "github.com/benbusby/b2"
	"os"
	"testing"
)

func TestListFiles(t *testing.T) {
	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	testFiles := [3]string{"a.txt", "b.txt", "c.txt"}
	for _, filename := range testFiles {
		uploadTestFile(filename)
	}

	test := func(service *Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		fullFileList, err := service.ListAllFiles(bucketID)
		if err != nil {
			t.Fatalf("Error listing all files: %v", err)
		} else if len(fullFileList.Files) < 3 {
			t.Fatal("Full file list does not contain all files")
		}

		partialFileList, err := service.ListNFiles(bucketID, 2)
		if err != nil {
			t.Fatalf("Error listing N files: %v", err)
		} else if len(partialFileList.Files) != 2 {
			t.Fatalf("Error: expected=%d, received=%d",
				2, len(partialFileList.Files))
		}
	}

	test(accountV2)
	test(accountV3)
}
