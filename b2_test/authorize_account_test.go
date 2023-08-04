package b2_test

import (
	. "github.com/benbusby/b2"
	"os"
	"reflect"
	"testing"
)

func TestAuthorizeAccount(t *testing.T) {
	bucketKeyID := os.Getenv("B2_TEST_KEY_ID")
	bucketKey := os.Getenv("B2_TEST_KEY")

	account, err := AuthorizeAccount(bucketKeyID, bucketKey)
	if err != nil {
		t.Fatal("Failed B2 account authorization")
	} else if reflect.ValueOf(account).IsZero() {
		t.Fatal("Empty response from B2")
	}
}
