package b2_test

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	. "github.com/benbusby/b2"
	"log"
	"os"
	"reflect"
	"testing"
)

const chunkSize = 5242880
const largeUploadSize = chunkSize + 15
const maxAttempts = 3

func uploadLargeFile(service Service) (LargeFile, error) {
	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	data := make([]byte, largeUploadSize)
	_, _ = rand.Read(data)

	var checksums []string
	filename := "large-file.txt"

	startFile, err := service.StartLargeFile(filename, bucketID)
	if err != nil {
		log.Printf("Failed to start large file: %v", err)
	}

	start := 0
	chunk := 1
	done := false
	for start < len(data) && !done {
		end := start + chunkSize

		if start+chunkSize > len(data) {
			end = len(data)
			done = true
		}

		contents := data[start:end]
		checksum := fmt.Sprintf("%x", sha1.Sum(contents))
		checksums = append(checksums, checksum)

		uploadChunk := func(attempt int) (bool, error) {
			if attempt > 0 {
				log.Printf("Attempt #%d", attempt+1)
			}

			partInfo, err := service.GetUploadPartURL(startFile.FileID)
			if err != nil {
				return false, err
			}

			err = UploadFilePart(partInfo, chunk, checksum, contents)
			if err != nil {
				return attempt < maxAttempts, err
			}

			return false, nil
		}

		var retry bool
		var uploadErr error
		attempt := 0
		for attempt < maxAttempts {
			retry, uploadErr = uploadChunk(attempt)
			if retry {
				attempt += 1
			} else if !retry || uploadErr == nil {
				break
			}
		}

		chunk += 1
		start += chunkSize
	}

	return service.FinishLargeFile(startFile.FileID, checksums)
}

func TestUploadLargeFile(t *testing.T) {
	test := func(service Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		largeFile, err := uploadLargeFile(service)

		if err != nil {
			t.Fatal("Failed to finish large file")
		} else if reflect.ValueOf(largeFile).IsZero() {
			t.Fatal("Empty large file response from B2")
		} else if largeFile.ContentLength != largeUploadSize {
			t.Fatalf("Content length does not match full upload size: "+
				"expected=%d, actual=%d", largeUploadSize, largeFile.ContentLength)
		}
	}

	test(accountV2)
	test(accountV3)
}

func TestCancelLargeFile(t *testing.T) {
	test := func(service Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		bucketID := os.Getenv("B2_TEST_BUCKET_ID")
		filename := "cancel-large-file.txt"

		startFile, err := service.StartLargeFile(filename, bucketID)
		if err != nil {
			t.Fatalf("Failed to start large file: %v", err)
		}

		canceled, err := service.CancelLargeFile(startFile.FileID)
		if err != nil || !canceled {
			t.Fatal("Failed to cancel large file")
		}
	}

	test(accountV2)
	test(accountV3)
}

func TestUploadLocalLargeFile(t *testing.T) {
	largeFile, err := uploadLargeFile(dummyAccount)
	if err != nil {
		t.Fatal("Failed to finish uploading local large file")
	} else if reflect.ValueOf(largeFile).IsZero() {
		t.Fatal("Empty large file response after writing to machine")
	} else if largeFile.ContentLength != largeUploadSize {
		t.Fatalf("Content length does not match full upload size: "+
			"expected=%d, actual=%d", largeUploadSize, largeFile.ContentLength)
	}
}
