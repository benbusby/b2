package b2_test

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	. "github.com/benbusby/b2"
	"os"
	"reflect"
	"strings"
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
	filename := "file.txt"

	file, err := UploadFile(info, filename, checksum, data)

	if err != nil {
		t.Fatal("Failed to upload file to B2")
	} else if reflect.ValueOf(file).IsZero() {
		t.Fatal("Empty response from B2")
	}
}

func TestUploadLocalFile(t *testing.T) {
	info, _ := dummyAccount.GetUploadURL("")

	data := make([]byte, 10)
	_, _ = rand.Read(data)

	checksum := ""
	filename := "local-file.txt"
	path := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(dummyAccount.LocalPath, "/"),
		filename)

	_, err := UploadFile(info, filename, checksum, data)

	if err != nil {
		t.Fatalf("Failed to \"upload\" file locally: %v", err)
	} else if _, err := os.Stat(path); err != nil {
		t.Fatal("Local file does not exist after writing")
	}
}
