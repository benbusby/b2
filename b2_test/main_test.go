package b2_test

import (
	"crypto/sha1"
	"fmt"
	. "github.com/benbusby/b2"
	"github.com/benbusby/b2/utils"
	"log"
	"os"
	"reflect"
	"testing"
)

var account Auth
var dummyAccount Auth

const localUploadsPath = "./test"
const testString = "lorem ipsum"

func TestMain(m *testing.M) {
	var err error

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
	dummyAccount, err = AuthorizeDummyAccount(localUploadsPath)
	if err != nil {
		log.Fatalf("Failed to setup dummy account")
	}

	//log.SetOutput(io.Discard)

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

	removed := 0
	for _, file := range files.Files {
		if !account.DeleteFile(file.FileID, file.FileName) {
			log.Printf("Failed to delete file %s (%s)\n",
				file.FileName,
				file.FileID)
		} else {
			removed += 1
		}
	}

	log.Printf("Removed %d test files from B2\n", removed)

	localFiles, err := dummyAccount.ListAllFiles("")
	if err != nil {
		log.Fatal("Unable to list local test files")
	}

	locallyRemoved := 0
	for _, file := range localFiles.Files {
		if !dummyAccount.DeleteFile(file.FileID, file.FileName) {
			log.Printf("Failed to delete local test file %s", file.FileName)
		} else {
			locallyRemoved += 1
		}
	}

	log.Printf("Removed %d local test files\n", locallyRemoved)
}

func TestLimitedDummyAccount(t *testing.T) {
	// Create dummy account with 1 byte storage limit
	limDumAcct, err := AuthorizeLimitedDummyAccount(localUploadsPath, 1)
	if err != nil {
		t.Fatal("Failed to set up limited dummy account")
	}

	info, _ := limDumAcct.GetUploadURL("")
	data := []byte(testString)
	checksum := ""
	filename := "too-big.txt"
	filepath := fmt.Sprintf("%s/%s", info.UploadURL, filename)

	_, err = UploadFile(info, filename, checksum, data)
	if err != utils.StorageError {
		log.Fatalf("Did not receive expected error: "+
			"expected=%v, actual=%v", utils.StorageError, err)
	} else if _, err := os.Stat(filepath); err == nil {
		log.Fatal("File should have failed to write, but was written " +
			"without any errors")
	}

}

func uploadTestFile(filename string) File {
	info, _ := account.GetUploadURL(os.Getenv("B2_TEST_BUCKET_ID"))
	data := []byte(testString)
	checksum := fmt.Sprintf("%x", sha1.Sum(data))
	file, _ := UploadFile(info, filename, checksum, data)

	return file
}

func uploadLocalTestFile(filename string) File {
	info, _ := dummyAccount.GetUploadURL("")
	data := []byte(testString)
	checksum := ""
	file, err := UploadFile(info, filename, checksum, data)
	if err != nil {
		log.Fatalf("Failed to create local file: %v", err)
	}

	return file
}
