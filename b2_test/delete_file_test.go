package b2_test

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	. "github.com/benbusby/b2"
	"os"
	"testing"
)

func TestDeleteFile(t *testing.T) {
	test := func(service *Service) {
		fmt.Printf("%s-- version %s\n", logPadding, service.APIVersion)
		info, _ := service.GetUploadURL(os.Getenv("B2_TEST_BUCKET_ID"))

		data := make([]byte, 10)
		_, _ = rand.Read(data)

		checksum := fmt.Sprintf("%x", sha1.Sum(data))
		filename := "delete-this.txt"

		file, _ := UploadFile(info, filename, checksum, data)

		deleted, err := service.DeleteFile(file.FileID, file.FileName)
		if !deleted || err != nil {
			t.Fatal("Failed to delete file from B2")
		}
	}

	test(accountV2)
	test(accountV3)
}

func TestDeleteLocalFile(t *testing.T) {
	file := uploadLocalTestFile("delete-this.txt")

	deleted, err := dummyAccount.DeleteFile(file.FileID, file.FileName)
	if !deleted || err != nil {
		t.Fatal("Failed to delete local file")
	}
}
