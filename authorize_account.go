package b2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/benbusby/b2/utils"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

const AuthURLV2 string = "https://api.backblazeb2.com/b2api/v2/b2_authorize_account"
const AuthURLV3 string = "https://api.backblazeb2.com/b2api/v3/b2_authorize_account"

type Service struct {
	APIURL             string
	AuthorizationToken string
	APIVersion         string
	Dummy              bool
	LocalPath          string
	StorageMaximum     int64
	Logging            bool
}

type AuthV2 struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountID               string `json:"accountId"`
	Allowed                 struct {
		BucketID     string   `json:"bucketId"`
		BucketName   string   `json:"bucketName"`
		Capabilities []string `json:"capabilities"`
		NamePrefix   any      `json:"namePrefix"`
	} `json:"allowed"`
	APIURL              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadURL         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
	S3APIURL            string `json:"s3ApiUrl"`
}

type AuthV3 struct {
	AccountID string `json:"accountId"`
	APIInfo   struct {
		StorageAPI struct {
			AbsoluteMinimumPartSize int      `json:"absoluteMinimumPartSize"`
			APIURL                  string   `json:"apiUrl"`
			BucketID                any      `json:"bucketId"`
			BucketName              any      `json:"bucketName"`
			Capabilities            []string `json:"capabilities"`
			DownloadURL             string   `json:"downloadUrl"`
			InfoType                string   `json:"infoType"`
			NamePrefix              any      `json:"namePrefix"`
			RecommendedPartSize     int      `json:"recommendedPartSize"`
			S3APIURL                string   `json:"s3ApiUrl"`
		} `json:"storageApi"`
		GroupsAPI struct {
			Capabilities []string `json:"capabilities"`
			GroupsAPIURL string   `json:"groupsApiUrl"`
			InfoType     string   `json:"infoType"`
		} `json:"groupsApi"`
	} `json:"apiInfo"`
	ApplicationKeyExpirationTimestamp any    `json:"applicationKeyExpirationTimestamp"`
	AuthorizationToken                string `json:"authorizationToken"`
}

func AuthorizeAccount(b2BucketKeyId, b2BucketKey string) (*Service, AuthV3, error) {
	response, err := InitAuthorization(b2BucketKeyId, b2BucketKey, AuthURLV3)
	if err != nil {
		return &Service{}, AuthV3{}, err
	}

	var auth AuthV3
	err = json.NewDecoder(response).Decode(&auth)
	if err != nil {
		return &Service{}, AuthV3{}, err
	}

	// Trim trailing slash
	apiURL := auth.APIInfo.StorageAPI.APIURL
	if strings.HasSuffix(apiURL, "/") {
		auth.APIInfo.StorageAPI.APIURL = apiURL[0 : len(apiURL)-2]
	}

	service := &Service{
		APIURL:             auth.APIInfo.StorageAPI.APIURL,
		AuthorizationToken: auth.AuthorizationToken,
		APIVersion:         "v3",
		Dummy:              false,
		LocalPath:          "",
		StorageMaximum:     0,
	}

	return service, auth, nil
}

func AuthorizeAccountV2(b2BucketKeyId, b2BucketKey string) (*Service, AuthV2, error) {
	response, err := InitAuthorization(b2BucketKeyId, b2BucketKey, AuthURLV2)
	if err != nil {
		return &Service{}, AuthV2{}, err
	}

	var auth AuthV2
	err = json.NewDecoder(response).Decode(&auth)
	if err != nil {
		return &Service{}, AuthV2{}, err
	}

	// Trim trailing slash
	if strings.HasSuffix(auth.APIURL, "/") {
		auth.APIURL = auth.APIURL[0 : len(auth.APIURL)-2]
	}

	service := &Service{
		APIURL:             auth.APIURL,
		AuthorizationToken: auth.AuthorizationToken,
		APIVersion:         "v2",
		Dummy:              false,
		LocalPath:          "",
		StorageMaximum:     0,
	}

	return service, auth, nil
}

func InitAuthorization(b2BucketKeyId, b2BucketKey, authURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return nil, err
	}

	authString := fmt.Sprintf("%s:%s", b2BucketKeyId, b2BucketKey)
	authString = base64.StdEncoding.EncodeToString([]byte(authString))

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Basic %s", authString)},
	}

	res, err := utils.Client.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode >= 400 {
		log.Printf("%s -- error: %d\n", authURL, res.StatusCode)
		resp, _ := httputil.DumpResponse(res, true)
		log.Println(fmt.Sprintf("%s", resp))
		return res.Body, utils.B2Error
	}

	return res.Body, nil
}

// AuthorizeDummyAccount allows using the B2 library as normal, but having
// all files saved and retrieved from a specific folder on the machine.
func AuthorizeDummyAccount(path string) (*Service, error) {
	if _, err := os.Stat(path); err != nil {
		// Attempt to create directory
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return &Service{}, err
		}
	}

	return &Service{
		Dummy:     true,
		LocalPath: path,
	}, nil
}

// AuthorizeLimitedDummyAccount functions the same as AuthorizeDummyAccount, but
// imposes an additional limitation for the total size of the directory specified
// in the "path" variable.
func AuthorizeLimitedDummyAccount(path string, storageLimit int64) (*Service, error) {
	service, err := AuthorizeDummyAccount(path)
	if err != nil {
		return &Service{}, err
	}

	service.StorageMaximum = storageLimit
	return service, nil
}

func (b2Service *Service) SetLogging(enable bool) {
	b2Service.Logging = enable
}

func (b2Service *Service) Logf(format string, v ...any) {
	if !b2Service.Logging {
		return
	}

	log.Printf(format, v)
}
