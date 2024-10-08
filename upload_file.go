package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/benbusby/b2/utils"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

const APIGetUploadURL string = "b2_get_upload_url"

// File represents the data returned by UploadFile
type File struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int64  `json:"contentLength"`
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

// FileInfo represents the data returned by GetUploadURL
type FileInfo struct {
	BucketID           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
	Dummy              bool
	StorageMaximum     int64
}

// GetUploadURL returns a FileInfo struct containing the URL to use
// for uploading a file, the ID of the bucket the file will be put
// in, and a token for authenticating the upload request.
func (b2Service *Service) GetUploadURL(bucketID string) (FileInfo, error) {
	if b2Service.Dummy {
		return FileInfo{
			UploadURL:      b2Service.LocalPath,
			StorageMaximum: b2Service.StorageMaximum,
			Dummy:          true,
		}, nil
	}

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIGetUploadURL)

	req, err := http.NewRequest("GET", reqURL, nil)

	q := req.URL.Query()
	q.Add("bucketId", bucketID)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		b2Service.Logf("B2Error creating new HTTP request: %v\n", err)
		return FileInfo{}, err
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		b2Service.Logf("B2Error requesting B2 upload URL: %v\n", err)
		return FileInfo{}, err
	} else if res.StatusCode >= 400 {
		b2Service.Logf("\n%s %s\n", "GET", reqURL)
		resp, _ := httputil.DumpResponse(res, true)
		return FileInfo{}, utils.NewB2Error(nil, string(resp))
	}

	var upload FileInfo
	err = json.NewDecoder(res.Body).Decode(&upload)
	if err != nil {
		b2Service.Logf("B2Error decoding B2 upload info: %v", err)
		return FileInfo{}, err
	}

	return upload, nil
}

// UploadFile uploads file byte content to B2 alongside a name for the file
// and a SHA1 checksum for the byte content. It returns a File object, which
// contains fields such as FileID and ContentLength which can be stored and
// used later to download the file.
func UploadFile(
	b2Info FileInfo,
	filename string,
	checksum string,
	contents []byte,
) (File, error) {
	if b2Info.Dummy {
		return uploadLocalFile(b2Info, filename, contents)
	}

	req, err := http.NewRequest(
		"POST",
		b2Info.UploadURL,
		bytes.NewBuffer(contents))
	if err != nil {
		return File{}, err
	}

	req.Header = http.Header{
		"Authorization":     {b2Info.AuthorizationToken},
		"Content-Type":      {"application/octet-stream"},
		"Content-Length":    {strconv.Itoa(len(contents))},
		"X-Bz-File-Name":    {filename},
		"X-Bz-Content-Sha1": {checksum},
	}

	res, err := utils.Client.Do(req)

	if err != nil {
		return File{}, err
	} else if res.StatusCode >= 400 {
		resp, _ := httputil.DumpResponse(res, true)
		return File{}, utils.NewB2Error(nil, string(resp))
	}

	var b2File File
	err = json.NewDecoder(res.Body).Decode(&b2File)
	if err != nil {
		return File{}, err
	}

	return b2File, nil
}

// uploadLocalFile skips the usual uploading to a B2 bucket and instead
// writes the file to a path specified in b2Info.UploadURL
func uploadLocalFile(
	b2Info FileInfo,
	filename string,
	contents []byte,
) (File, error) {
	if _, err := os.Stat(b2Info.UploadURL); err != nil {
		return File{}, err
	}

	if b2Info.StorageMaximum > 0 {
		dirSize, err := utils.CheckDirSize(b2Info.UploadURL)
		if err != nil {
			return File{}, err
		}

		if dirSize+int64(len(contents)) > int64(b2Info.StorageMaximum) {
			return File{}, utils.StorageError
		}
	}

	path := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(b2Info.UploadURL, "/"),
		filename)

	file, err := os.Create(path)
	if err != nil {
		return File{}, err
	}

	_, err = file.Write(contents)
	if err != nil {
		return File{}, err
	}

	return File{
		FileID:        filename,
		BucketID:      filename,
		FileName:      filename,
		ContentLength: int64(len(contents)),
	}, nil
}
