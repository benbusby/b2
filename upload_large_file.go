package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"yeetfile/b2/utils"
)

const APIStartLargeFile string = "b2_start_large_file"
const APIGetUploadPartURL string = "b2_get_upload_part_url"
const APIFinishLargeFile = "b2_finish_large_file"

// StartFile represents the data returned by StartLargeFile
type StartFile struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileID        string `json:"fileId"`
	FileInfo      struct {
	} `json:"fileInfo"`
	FileName      string `json:"fileName"`
	FileRetention struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    struct {
			Mode                 any `json:"mode"`
			RetainUntilTimestamp any `json:"retainUntilTimestamp"`
		} `json:"value"`
	} `json:"fileRetention"`
	LegalHold struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    any  `json:"value"`
	} `json:"legalHold"`
	ServerSideEncryption struct {
		Algorithm any `json:"algorithm"`
		Mode      any `json:"mode"`
	} `json:"serverSideEncryption"`
	UploadTimestamp int64 `json:"uploadTimestamp"`
}

// FilePartInfo represents the data returned by GetUploadPartURL
type FilePartInfo struct {
	FileID             string `json:"fileId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

// LargeFile represents the file object created by FinishLargeFile
type LargeFile struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentMd5    any    `json:"contentMd5"`
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

// StartLargeFile begins the process for uploading a multi-chunk file to B2.
// The filename provided cannot change once the large file upload has begun.
func (b2Auth Auth) StartLargeFile(
	filename string,
) (StartFile, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"bucketId": "%s",
		"fileName": "%s",
		"contentType": "b2/x-auto"
	}`, os.Getenv("B2_BUCKET_ID"), filename)))
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, utils.APIPrefix, APIStartLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return StartFile{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("Error starting B2 file: %v\n", err)
		return StartFile{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return StartFile{}, utils.Error
	}

	var file StartFile
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		log.Printf("Error decoding B2 file init: %v", err)
		return StartFile{}, err
	}

	return file, nil
}

// GetUploadPartURL generates a URL and token for uploading individual chunks
// of a file to B2. It requires a StartFile struct returned by StartLargeFile,
// which contains the unique file ID for this new file.
func (b2Auth Auth) GetUploadPartURL(
	b2File StartFile,
) (FilePartInfo, error) {
	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, utils.APIPrefix, APIGetUploadPartURL)

	req, err := http.NewRequest("GET", reqURL, nil)

	q := req.URL.Query()
	q.Add("fileId", b2File.FileID)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return FilePartInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("Error getting B2 upload url: %v\n", err)
		return FilePartInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "GET", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return FilePartInfo{}, utils.Error
	}

	var upload FilePartInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("Error decoding B2 upload part info: %v", err)
		return FilePartInfo{}, err
	}

	return upload, nil
}

// UploadFilePart uploads a single chunk of file data to the URL provided by
// GetUploadPartURL. Each subsequent chunk should increment chunkNum, with the
// first chunk starting at 1 (not 0). Each chunk should be provided with a
// SHA1 checksum as well.
func (b2PartInfo FilePartInfo) UploadFilePart(
	chunkNum int,
	checksum string,
	contents []byte,
) error {
	req, err := http.NewRequest(
		"POST",
		b2PartInfo.UploadURL,
		bytes.NewBuffer(contents))
	if err != nil {
		log.Printf("Error creating upload request: %v\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization":     {b2PartInfo.AuthorizationToken},
		"Content-Length":    {strconv.Itoa(len(contents))},
		"X-Bz-Part-Number":  {strconv.Itoa(chunkNum)},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := utils.Client.Do(req)

	if err != nil {
		log.Printf("Error uploading file to B2: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", b2PartInfo.UploadURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return utils.Error
	}

	return nil
}

// FinishLargeFile completes the chunked upload process. The FileID from
// calling StartLargeFile should be used here, and all checksums from
// UploadFilePart should be passed a string-ified array.
// For example: "['checksum1', 'checksum2']"
func (b2Auth Auth) FinishLargeFile(
	fileID string,
	checksums string,
) (LargeFile, error) {
	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"partSha1Array": %s
	}`, fileID, checksums)))

	reqURL := fmt.Sprintf(
		"%s/%s/%s",
		b2Auth.APIURL, utils.APIPrefix, APIFinishLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("Error creating new HTTP request: %v\n", err)
		return LargeFile{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Auth.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)

	if err != nil {
		log.Printf("Error finishing B2 upload: %v\n", err)
		return LargeFile{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		fmt.Println(fmt.Sprintf("%s", resp))
		return LargeFile{}, utils.Error
	}

	var largeFile LargeFile
	err = json.NewDecoder(res.Body).Decode(&largeFile)
	if err != nil {
		log.Printf("Error decoding B2 large file info: %v", err)
		return LargeFile{}, err
	}

	return largeFile, nil
}
