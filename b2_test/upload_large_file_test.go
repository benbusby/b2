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

func uploadLargeFile(auth Auth) (LargeFile, error) {
	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	data := make([]byte, largeUploadSize)
	_, _ = rand.Read(data)

	var checksums []string
	filename := "large-file.txt"

	startFile, err := auth.StartLargeFile(filename, bucketID)
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

		partInfo, err := auth.GetUploadPartURL(startFile)
		if err != nil {
			log.Print("Failed to get upload part url")
		}

		err = UploadFilePart(partInfo, chunk, checksum, contents)
		if err != nil {
			log.Print("Failed to upload file chunk")
		}

		chunk += 1
		start += chunkSize
	}

	return auth.FinishLargeFile(startFile.FileID, checksums)
}

func TestUploadLargeFile(t *testing.T) {
	largeFile, err := uploadLargeFile(account)
	if err != nil {
		t.Fatal("Failed to finish large file")
	} else if reflect.ValueOf(largeFile).IsZero() {
		t.Fatal("Empty large file response from B2")
	} else if largeFile.ContentLength != largeUploadSize {
		t.Fatalf("Content length does not match full upload size: "+
			"expected=%d, actual=%d", largeUploadSize, largeFile.ContentLength)
	}
}

func TestCancelLargeFile(t *testing.T) {
	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	filename := "cancel-large-file.txt"

	startFile, err := account.StartLargeFile(filename, bucketID)
	if err != nil {
		t.Fatalf("Failed to start large file: %v", err)
	}

	canceled, err := account.CancelLargeFile(startFile.FileID)
	if err != nil || !canceled {
		t.Fatal("Failed to cancel large file")
	}
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
