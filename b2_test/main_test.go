package b2

import (
	"io"
	"log"
	"os"
	"testing"
)

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

	log.SetOutput(io.Discard)

	code := m.Run()
	cleanup()
	os.Exit(code)
}

// cleanup removes all files from the B2 test bucket
func cleanup() {

}
