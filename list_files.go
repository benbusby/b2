package b2

import (
	"encoding/json"
	"fmt"
	"github.com/benbusby/b2/utils"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const APIListFileVersions = "b2_list_file_versions"

type FileListItem struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int64  `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentMd5    string `json:"contentMd5"`
	ContentType   string `json:"contentType"`
	FileID        string `json:"fileId"`
	FileInfo      struct {
		SrcLastModifiedMillis string `json:"src_last_modified_millis"`
	} `json:"fileInfo"`
	FileName      string `json:"fileName"`
	FileRetention struct {
		IsClientAuthorizedToRead bool `json:"isClientAuthorizedToRead"`
		Value                    struct {
			Mode                 string `json:"mode"`
			RetainUntilTimestamp string `json:"retainUntilTimestamp"`
		} `json:"value"`
	} `json:"fileRetention"`
	LegalHold struct {
		IsClientAuthorizedToRead bool   `json:"isClientAuthorizedToRead"`
		Value                    string `json:"value"`
	} `json:"legalHold"`
	ReplicationStatus    string `json:"replicationStatus"`
	ServerSideEncryption struct {
		Algorithm string `json:"algorithm"`
		Mode      string `json:"mode"`
	} `json:"serverSideEncryption"`
	UploadTimestamp int `json:"uploadTimestamp"`
}

type FileList struct {
	Files        []FileListItem `json:"files"`
	NextFileName string         `json:"nextFileName"`
	NextFileID   string         `json:"nextFileId"`
}

// ListAllFiles is a helper function for simply fetching all available files in
// the bucket. If more than 100 files exist, the FileList struct will contain
// NextFileName and NextFileID fields that can be used with ListFiles to fetch
// the remainder.
func (b2Service *Service) ListAllFiles(bucketID string) (FileList, error) {
	return b2Service.ListFiles(bucketID, 100, "", "")
}

// ListNFiles is similar to ListAllFiles, but allows explicitly stating how many
// files you want returned in the response.
func (b2Service *Service) ListNFiles(bucketID string, count int) (FileList, error) {
	return b2Service.ListFiles(bucketID, count, "", "")
}

// ListFiles lists all files in the specified bucket up to a maximum of `count`,
// starting with `startName` and, optionally, `startID`. If count is set to an
// invalid or negative value, the default number of files returned is 100. If
// startName or startID are not set, the bucket will list all files
func (b2Service *Service) ListFiles(
	bucketID string,
	count int,
	startName string,
	startID string,
) (FileList, error) {
	if b2Service.Dummy {
		return listLocalFiles(b2Service.LocalPath)
	}

	reqURL := utils.FormatB2URL(
		b2Service.APIURL, b2Service.APIVersion, APIListFileVersions)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return FileList{}, err
	}

	q := req.URL.Query()
	q.Add("bucketId", bucketID)
	q.Add("maxFileCount", fmt.Sprintf("%d", count))

	if len(startName) > 0 {
		q.Add("startFileName", startName)
	}

	if len(startID) > 0 {
		q.Add("startFileId", startID)
	}

	req.URL.RawQuery = q.Encode()
	req.Header = http.Header{
		"Authorization": {b2Service.AuthorizationToken},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		b2Service.Logf("B2Error requesting B2 file list: %v\n", err)
		return FileList{}, err
	} else if res.StatusCode >= 400 {
		resp, _ := httputil.DumpResponse(res, true)
		return FileList{}, utils.NewB2Error(nil, string(resp))
	}

	var b2FileList FileList
	err = json.NewDecoder(res.Body).Decode(&b2FileList)
	if err != nil {
		b2Service.Logf("B2Error decoding B2 file list: %v", err)
		return FileList{}, err
	}

	return b2FileList, nil
}

// listLocalFiles returns all files within the specified path. Unlike the
// B2 version of listing files, listing local files will return all files
// within the directory.
func listLocalFiles(path string) (FileList, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return FileList{}, err
	}

	var fileList []FileListItem
	for _, file := range dir {
		name := file.Name()
		filePath := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), name)
		stat, err := os.Stat(filePath)
		if err != nil {
			return FileList{}, err
		}

		fileList = append(fileList, FileListItem{
			FileName:      name,
			FileID:        name,
			BucketID:      name,
			ContentLength: stat.Size(),
		})
	}

	return FileList{Files: fileList}, nil
}
