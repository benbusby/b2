package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"yeetfile/src/b2"
)

const APIDownloadById string = "b2_download_file_by_id"

func setupDownload(apiURL string, fileID string) (*http.Request, error) {
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		apiURL, b2.APIPrefix, APIDownloadById)

	req, err := http.NewRequest("GET", reqURL, nil)

	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("fileId", fileID)
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func download(req *http.Request) ([]byte, error) {
	res, err := b2.B2Client.Do(req)
	if err != nil {
		log.Printf("Error requesting B2 download: %v\n", err)
		return nil, err
	} else if res.StatusCode >= 400 {
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return nil, b2.B2Error
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error reading response body")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	return body, nil
}

func (b2Auth Auth) PartialDownloadById(
	id string,
	begin int,
	end int,
) ([]byte, error) {
	req, err := setupDownload(b2Auth.APIURL, id)
	if err != nil {
		log.Fatalf("Error setting up download: %v", err)
		return nil, err
	}

	byteRange := fmt.Sprintf("bytes=%d-%d", begin, end)

	req.Header = http.Header{
		"Authorization": {b2Auth.AuthorizationToken},
		"Range":         {byteRange},
	}

	return download(req)
}

func (b2Auth Auth) DownloadById(id string) ([]byte, error) {
	req, err := setupDownload(b2Auth.APIURL, id)
	if err != nil {
		log.Fatalf("Error setting up download: %v", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {b2Auth.AuthorizationToken},
	}

	return download(req)
}
