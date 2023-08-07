package b2

import (
	. "github.com/benbusby/b2"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
)

var account Auth

func TestMain(m *testing.M) {
	// Ensure all required environment variables have been set
	// before running tests
	if len(os.Getenv("B2_TEST_KEY_ID")) == 0 {
		log.Fatal("--- missing B2_TEST_KEY_ID")
	} else if len(os.Getenv("B2_TEST_KEY")) == 0 {
		log.Fatal("--- missing B2_TEST_KEY")
	} else if len(os.Getenv("B2_TEST_BUCKET_ID")) == 0 {
		log.Fatal("--- missing B2_TEST_BUCKET_ID")
	}

	account = authorizeAccount()

	log.SetOutput(io.Discard)

	code := m.Run()
	cleanup()
	os.Exit(code)
}

// authorizeAccount sets up authorization with B2, which is a prerequisite for
// testing B2 functionality.
func authorizeAccount() Auth {
	bucketKeyID := os.Getenv("B2_TEST_KEY_ID")
	bucketKey := os.Getenv("B2_TEST_KEY")

	b2Account, err := AuthorizeAccount(bucketKeyID, bucketKey)
	if err != nil {
		log.Fatal("Unable to authorize B2 account")
	} else if reflect.ValueOf(b2Account).IsZero() {
		log.Fatal("Empty authorization response from B2")
	}

	return b2Account
}

// cleanup removes all files from the B2 test bucket
func cleanup() {
	log.SetOutput(os.Stderr)

	bucketID := os.Getenv("B2_TEST_BUCKET_ID")
	files, err := account.ListAllFiles(bucketID)
	if err != nil {
		log.Fatal("Unable to clean up testing files")
	}

	for _, file := range files.Files {
		if !account.DeleteFile(file.FileID, file.FileName) {
			log.Printf("Failed to delete file %s (%s)\n",
				file.FileName,
				file.FileID)
		}
	}
}
