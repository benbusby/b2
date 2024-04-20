package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/benbusby/b2/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

const APIStartLargeFile string = "b2_start_large_file"
const APIGetUploadPartURL string = "b2_get_upload_part_url"
const APIFinishLargeFile = "b2_finish_large_file"
const APICancelLargeFile = "b2_cancel_large_file"

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
	Dummy              bool
	StorageMaximum     int
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
func (b2Service Service) StartLargeFile(
	filename string,
	bucketID string,
) (StartFile, error) {
	if b2Service.Dummy {
		return StartFile{FileID: filename, FileName: filename}, nil
	}

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"bucketId": "%s",
		"fileName": "%s",
		"contentType": "b2/x-auto"
	}`, bucketID, filename)))
	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIStartLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return StartFile{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("B2Error starting B2 file: %v\n", err)
		return StartFile{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return StartFile{}, utils.B2Error
	}

	var file StartFile
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		log.Printf("B2Error decoding B2 file init: %v", err)
		return StartFile{}, err
	}

	return file, nil
}

// GetUploadPartURL generates a URL and token for uploading individual chunks
// of a file to B2. It requires a StartFile struct returned by StartLargeFile,
// which contains the unique file ID for this new file.
func (b2Service Service) GetUploadPartURL(fileID string) (FilePartInfo, error) {
	if b2Service.Dummy {
		return FilePartInfo{
			FileID:         fileID,
			UploadURL:      b2Service.LocalPath,
			Dummy:          true,
			StorageMaximum: b2Service.StorageMaximum,
		}, nil
	}

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIGetUploadPartURL)

	req, err := http.NewRequest("GET", reqURL, nil)

	q := req.URL.Query()
	q.Add("fileId", fileID)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return FilePartInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		log.Printf("B2Error getting B2 upload url: %v\n", err)
		return FilePartInfo{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "GET", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return FilePartInfo{}, utils.B2Error
	}

	var upload FilePartInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		log.Printf("B2Error decoding B2 upload part info: %v", err)
		return FilePartInfo{}, err
	}

	return upload, nil
}

// UploadFilePart uploads a single chunk of file data to the URL provided by
// GetUploadPartURL. Each subsequent chunk should increment chunkNum, with the
// first chunk starting at 1 (not 0). Each chunk should be provided with a
// SHA1 checksum as well.
func UploadFilePart(
	b2PartInfo FilePartInfo,
	chunkNum int,
	checksum string,
	contents []byte,
) error {
	if b2PartInfo.Dummy {
		return uploadLocalFilePart(b2PartInfo, contents)
	}

	req, err := http.NewRequest(
		"POST",
		b2PartInfo.UploadURL,
		bytes.NewBuffer(contents))
	if err != nil {
		log.Printf("B2Error creating upload request: %v\n", err)
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
		log.Printf("B2Error uploading file to B2: %v\n", err)
		return err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", b2PartInfo.UploadURL)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return utils.B2Error
	}

	return nil
}

// CancelLargeFile cancels an in-progress large file upload and deletes the
// partial file from the B2 bucket. Returns true if the file was successfully
// deleted, otherwise false.
// Requires the fileID returned from StartLargeFile.
func (b2Service Service) CancelLargeFile(fileID string) (bool, error) {
	if b2Service.Dummy {
		return cancelLocalLargeFile(fileID, b2Service.LocalPath)
	}

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s"
	}`, fileID)))

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APICancelLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return false, err
	}

	req.Header = http.Header{
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)

	if err != nil {
		log.Printf("B2Error canceling B2 large file: %v\n", err)
		return false, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return false, utils.B2Error
	}

	return true, nil
}

// FinishLargeFile completes the chunked upload process. The FileID from
// calling StartLargeFile should be used here, and all checksums from
// UploadFilePart should be passed a string-ified array.
// For example: "['checksum1', 'checksum2']"
func (b2Service Service) FinishLargeFile(
	fileID string,
	checksums []string,
) (LargeFile, error) {
	if b2Service.Dummy {
		return finishLargeLocalFile(fileID, b2Service.LocalPath)
	}

	checksumsString := "[\"" + strings.Join(checksums, "\",\"") + "\"]"

	reqBody := bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"fileId": "%s",
		"partSha1Array": %s
	}`, fileID, checksumsString)))

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIFinishLargeFile)

	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		log.Printf("B2Error creating new HTTP request: %v\n", err)
		return LargeFile{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)

	if err != nil {
		log.Printf("B2Error finishing B2 upload: %v\n", err)
		return LargeFile{}, err
	} else if res.StatusCode >= 400 {
		log.Printf("\n%s %s\n", "POST", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return LargeFile{}, utils.B2Error
	}

	var largeFile LargeFile
	err = json.NewDecoder(res.Body).Decode(&largeFile)
	if err != nil {
		log.Printf("B2Error decoding B2 large file info: %v", err)
		return LargeFile{}, err
	}

	return largeFile, nil
}

// uploadLocalFilePart writes part of a file to the machine instead of to a B2
// bucket
func uploadLocalFilePart(info FilePartInfo, contents []byte) error {
	if info.StorageMaximum > 0 {
		dirSize, err := utils.CheckDirSize(info.UploadURL)
		if err != nil {
			return err
		}

		if dirSize+int64(len(contents)) > int64(info.StorageMaximum) {
			success, _ := cancelLocalLargeFile(info.FileID, info.UploadURL)
			if !success {
				log.Printf("Failed to cancel local file that " +
					"would hav exceeded the specified max " +
					"storage")
			}
			return utils.StorageError
		}
	}

	filename := fmt.Sprintf("%s/%s", strings.TrimSuffix(info.UploadURL, "/"), info.FileID)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Failed to close partial local file")
		}
	}(f)

	_, err = f.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

// cancelLocalLargeFile cancels an in-progress large file being written to
// disk by deleting it.
func cancelLocalLargeFile(id string, path string) (bool, error) {
	if len(id) == 0 {
		log.Println("Skipping attempt to cancel a large file upload " +
			"with no id")
		return false, nil
	}

	return deleteLocalFile(id, path), nil
}

// finishLargeLocalFile completes the process of uploading a file chunk-by-chunk
// to the local machine
func finishLargeLocalFile(id string, path string) (LargeFile, error) {
	filePath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), id)
	stat, err := os.Stat(filePath)
	if err != nil {
		return LargeFile{}, err
	}

	return LargeFile{
		FileID:        id,
		FileName:      id,
		BucketID:      id,
		ContentLength: int(stat.Size()),
	}, nil
}
