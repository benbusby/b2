package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const APIPrefix string = "b2api"

var Client = &http.Client{Timeout: 10 * time.Second}
var B2Error = errors.New("b2 client error")
var StorageError = errors.New("local storage has been exceeded")

func CheckDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func FormatB2URL(apiURL, apiVersion, endpoint string) string {
	return fmt.Sprintf(
		"%s/%s/%s/%s",
		apiURL, APIPrefix, apiVersion, endpoint)
}

func NewB2Error(err error, errMsg string) error {
	fullMsg := fmt.Sprintf("B2 Error: %v\n%s", err, errMsg)
	return errors.New(fullMsg)
}
