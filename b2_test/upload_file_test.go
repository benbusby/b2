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

	test := func(service *Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		info, err := service.GetUploadURL(bucketID)

		if err != nil {
			t.Fatal("Failed to get upload url from B2")
		} else if reflect.ValueOf(info).IsZero() {
			t.Fatal("Empty response from B2")
		}
	}

	test(accountV2)
	test(accountV3)
}

func TestUploadFile(t *testing.T) {
	data := make([]byte, 10)
	_, _ = rand.Read(data)

	checksum := fmt.Sprintf("%x", sha1.Sum(data))
	filename := "file.txt"

	test := func(service *Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		info, _ := service.GetUploadURL(os.Getenv("B2_TEST_BUCKET_ID"))

		file, err := UploadFile(info, filename, checksum, data)

		if err != nil {
			t.Fatal("Failed to upload file to B2")
		} else if reflect.ValueOf(file).IsZero() {
			t.Fatal("Empty response from B2")
		}
	}

	test(accountV2)
	test(accountV3)
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
