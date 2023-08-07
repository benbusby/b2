package b2

import (
	"bytes"
	"fmt"
	"github.com/benbusby/b2/utils"
	"log"
	"net/http"
	"net/http/httputil"
)

const APIDeleteFile = "b2_delete_file_version"

// DeleteFile removes a file from B2 using the file's ID and name. Both fields
// are required, and are provided when a file finishes uploading.
func (b2Auth Auth) DeleteFile(b2ID string, name string) bool {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"fileName": "%s"
	}`, b2ID, name)))

	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, utils.APIPrefix, APIDeleteFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return false
	}

	req.Header = http.Header{
		"Authorization": {b2Auth.AuthorizationToken},
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
