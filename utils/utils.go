package utils

import (
	"errors"
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
