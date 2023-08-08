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
	info, _ := account.GetUploadURL(os.Getenv("B2_TEST_BUCKET_ID"))

	data := make([]byte, 10)
	_, _ = rand.Read(data)

	checksum := fmt.Sprintf("%x", sha1.Sum(data))
	filename := "delete-this.txt"

	file, _ := UploadFile(info, filename, checksum, data)

	if !account.DeleteFile(file.FileID, file.FileName) {
		t.Fatal("Failed to delete file from B2")
	}
}
