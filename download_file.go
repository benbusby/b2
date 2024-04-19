package b2

import (
	"fmt"
	"github.com/benbusby/b2/utils"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const APIDownloadById string = "b2_download_file_by_id"

// setupDownload creates an http.Request with the URL for downloading a file,
// as well as the file ID included in the query.
func setupDownload(apiURL, apiVersion, fileID string) (*http.Request, error) {
	reqURL := fmt.Sprintf(
		"%s/%s/%s/%s",
		apiURL, utils.APIPrefix, apiVersion, APIDownloadById)

	req, err := http.NewRequest("GET", reqURL, nil)

	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("fileId", fileID)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

// download uses the http.Request returned by setupDownload to execute the
// request and return the []byte file content from B2.
func download(req *http.Request) ([]byte, error) {
	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("B2Error requesting B2 download: %v\n", err)
		return nil, err
	} else if res.StatusCode >= 400 {
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return nil, utils.B2Error
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("B2Error reading response body")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// PartialDownloadById downloads a file from B2 with a specified begin and end
// byte. For example, setting begin to 0 and end to 99 will download only the
// first 99 bytes of the file.
func (b2Service Service) PartialDownloadById(
	id string,
	begin int,
	end int,
) ([]byte, error) {
	if b2Service.Dummy {
		return partiallyDownloadLocalFile(
			id,
			b2Service.LocalPath,
			begin,
			end)
	}

	req, err := setupDownload(b2Service.APIURL, b2Service.APIVersion, id)
	if err != nil {
		log.Fatalf("B2Error setting up download: %v", err)
		return nil, err
	}

	byteRange := fmt.Sprintf("bytes=%d-%d", begin, end)

	req.Header = http.Header{
		"Authorization": {b2Service.AuthorizationToken},
		"Range":         {byteRange},
	}

	return download(req)
}

// DownloadById downloads an entire file (regardless of size) from B2.
func (b2Service Service) DownloadById(id string) ([]byte, error) {
	if b2Service.Dummy {
		return downloadLocalFile(id, b2Service.LocalPath)
	}

	req, err := setupDownload(b2Service.APIURL, b2Service.APIVersion, id)
	if err != nil {
		log.Fatalf("B2Error setting up download: %v", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {b2Service.AuthorizationToken},
	}

	return download(req)
}

// downloadLocalFile "downloads" a local file from the specified path + ID
// rather than fetching from B2.
func downloadLocalFile(id string, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), id)
	return os.ReadFile(fullPath)
}

// partiallyDownloadLocalFile retrieves a portion of a local file rather than
// fetching it from B2.
func partiallyDownloadLocalFile(
	id string,
	path string,
	begin int,
	end int,
) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), id)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	// B2 downloads encapsulate the end byte as well, whereas local reads
	// stop at the end byte. Modifying the end by +1 accounts for this
	// difference in order to get the download behavior to act the same.
	end += 1

	contents := make([]byte, end-begin)
	_, err = file.ReadAt(contents, int64(begin))
	if err != nil {
		return nil, err
	}

	return contents, err
}
