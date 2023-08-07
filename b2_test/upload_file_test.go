package b2

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	. "github.com/benbusby/b2"
	"os"
	"reflect"
	"testing"
)

func TestGetUploadURL(t *testing.T) {
	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	info, err := account.GetUploadURL(bucketID)

	if err != nil {
		t.Fatal("Failed to get upload url from B2")
	} else if reflect.ValueOf(info).IsZero() {
		t.Fatal("Empty response from B2")
	}
}

func TestUploadFile(t *testing.T) {
	info, _ := account.GetUploadURL(os.Getenv("B2_TEST_BUCKET_ID"))

	data := make([]byte, 10)
	_, _ = rand.Read(data)

	checksum := fmt.Sprintf("%x", sha1.Sum(data))
	filename := fmt.Sprintf("%s.txt", checksum)

	file, err := UploadFile(info, filename, checksum, data)

	if err != nil {
		t.Fatal("Failed to upload file to B2")
	} else if reflect.ValueOf(file).IsZero() {
		t.Fatal("Empty response from B2")
	}
}
