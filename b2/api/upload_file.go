package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"yeetfile/src/b2"
)

const APIGetUploadURL string = "b2_get_upload_url"

type File struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentMd5    string `json:"contentMd5"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileID        string `json:"fileId"`
	FileInfo      struct {
	} `json:"fileInfo"`
	FileName      string `json:"fileName"`
	FileRetention struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    any  `json:"value"`
	} `json:"fileRetention"`
	LegalHold struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    any  `json:"value"`
	} `json:"legalHold"`
	ServerSideEncryption struct {
		Algorithm string `json:"algorithm"`
		Mode      string `json:"mode"`
	} `json:"serverSideEncryption"`
	UploadTimestamp int64 `json:"uploadTimestamp"`
}

type FileInfo struct {
	BucketID           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

func (b2Auth Auth) GetUploadURL() (FileInfo, error) {
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, b2.APIPrefix, APIGetUploadURL)

	req, err := http.NewRequest("GET", reqURL, nil)

	q := req.URL.Query()
	q.Add("bucketId", os.Getenv("B2_BUCKET_ID"))
	req.URL.RawQuery = q.Encode()

	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return FileInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := b2.B2Client.Do(req)
	if err != nil {
		log.Printf("Error requesting B2 upload URL: %v\n", err)
		return FileInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "GET", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return FileInfo{}, b2.B2Error
	}

	var upload FileInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload info: %v", err)
		return FileInfo{}, err
	}

	return upload, nil
}

func (b2Info FileInfo) UploadFile(
	filename string,
	checksum string,
	contents []byte,
) (File, error) {
	req, err := http.NewRequest(
		"POST",
		b2Info.UploadURL,
		bytes.NewBuffer(contents))
	if err != nil {
		log.Printf("Error creating upload request: %v\n", err)
		return File{}, err
	}

	req.Header = http.Header{
		"Authorization":     {b2Info.AuthorizationToken},
		"Content-Type":      {"application/octet-stream"},
		"Content-Length":    {strconv.Itoa(len(contents))},
		"X-Bz-File-Name":    {filename},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := b2.B2Client.Do(req)

	if err != nil {
		log.Printf("Error uploading file chunk to B2: %v\n", err)
		return File{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", b2Info.UploadURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return File{}, b2.B2Error
	}

	var b2File File
	err = json.NewDecoder(res.Body).Decode(&b2File)
	if err != nil {
		log.Printf("Error decoding B2 file: %v", err)
		return File{}, err
	}

	return b2File, nil
}
