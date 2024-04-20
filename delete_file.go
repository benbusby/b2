package b2

import (
	"bytes"
	"fmt"
	"github.com/benbusby/b2/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const APIDeleteFile = "b2_delete_file_version"

// DeleteFile removes a file from B2 using the file's ID and name. Both fields
// are required, and are provided when a file finishes uploading.
func (b2Service Service) DeleteFile(b2ID string, name string) bool {
	if b2Service.Dummy {
		return deleteLocalFile(b2ID, b2Service.LocalPath)
	}

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"fileName": "%s"
	}`, b2ID, name)))

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIDeleteFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return false
	}

	req.Header = http.Header{
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("%s error: %v\n", APIDeleteFile, err)
		return false
	} else if res.StatusCode >= 400 {
		log.Printf("%s err: %d\n", APIDeleteFile, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return false
	}

	return true
}

// deleteLocalFile removes a file from the local machine
func deleteLocalFile(id string, path string) bool {
	if len(id) == 0 {
		log.Println("Attempting to delete without specifying an id, skipping")
		return false
	}

	fullPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), id)
	if err := os.Remove(fullPath); err != nil {
		return false
	}

	return true
}
