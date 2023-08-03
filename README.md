# b2
A Go library for the [Backblaze B2 Cloud Storage
](https://www.backblaze.com/b2/cloud-storage.html) API.

## Contents

1. [API Support](#api-support)
2. [Usage](#usage)
   1. [Authentication](#authentication)
   2. [Upload File](#upload-file)
   3. [Upload Large File](#upload-large-file)
   4. [Download File](#download-file)
   5. [Delete a File](#delete-a-file)

## API Support

The following API endpoints and functionality are currently supported:

- Authentication
  - `b2_authorize_account`
- Uploading a file
  - `b2_get_upload_url`
  - `b2_upload_file`
- Uploading a large file (multi-part upload)
  - `b2_start_large_file`
  - `b2_get_upload_part_url`
  - `b2_upload_part`
  - `b2_finish_large_file`
  - `b2_cancel_large_file`
- Downloading a file
  - `b2_download_file_by_id`
- Deleting a file
  - `b2_delete_file_version`
 
The project is being actively developed, and more functionaly will likely
be added in the near future. Existing functionality is unlikely to change
and should be considered stable.

## Usage

### Authentication

You can authenticate with Backblaze B2 using `b2.AuthorizeAccount(keyID, key)`,
where `keyID` is either an application or master key ID, and `key` is the
actual key contents. `b2.AuthorizeAccount` returns a `b2.Auth` struct:

```go
type Auth struct {
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
```

Most B2 functions have a receiver type of `Auth` and will use the
`AuthorizationToken` field to authenticate requests.

___

#### Function

```go
func AuthorizeAccount(
	b2BucketKeyId string,
	b2BucketKey string,
) (Auth, error)
```

___

#### Example

```go
b2, err := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

if err != nil {
	panic(err)
}
```

### Upload File

Uploading a regular (non-chunked) file involves fetching the upload
parameters first (unique upload URL, token, etc), and then sending
your file data, a SHA-1 checksum for the data, and the file name.

Note that although the endpoint for retrieving upload data is named
`b2_get_upload_url`, the returned struct contains more than just a
URL. Most importantly, it contains the required auth token for actually
uploading the file.

After uploading, you'll receive a struct with fields such as `FileID` that
you can use to access or delete the file later.

___

#### Functions

Get upload URL:
```go
func (b2Auth Auth) GetUploadURL() (FileInfo, error)
```

Upload file:
```go
func (b2Info FileInfo) UploadFile(
	filename string,
	checksum string,
	contents []byte,
) (File, error)
```

___

#### Example

```go
b2, _ := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

b2Uploader, _ := b2.GetUploadURL()

data := []byte("test")

h := sha1.New()
h.Write(data)
checksum := fmt.Sprintf("%x", h.Sum(nil))

file, err := b2Uploader.UploadFile(
	"myfile.txt",
	checksum,
	data)

// save/store `file.FileID` somewhere in order to access it later
```

### Upload Large File

Uploading a large file requires extra steps to "start" and "stop" uploading,
which will depend on how many chunks of data you're sending. Each chunk of
file data needs to be at least 5mb, except for the final chunk. You cannot use
the large file upload process to upload files <5mb.

The basic flow of uploading a large file is:

1. Start the file
2. Get the upload info for a chunk
3. Upload the chunks of data until finished
4. Stop the file

You'll also need to track each checksum as you upload data, since finishing a
large file requires an array of past checksums to finalize the upload.

The finalized large file struct, like the normal B2 file struct, contains metadata
that you may want in order to access the file later.

___

#### Function(s)

Start large file:
```go
func (b2Auth Auth) StartLargeFile(filename string) (StartFile, error)
```

Get upload part URL:
```go
func (b2Auth Auth) GetUploadPartURL(b2File StartFile) (FilePartInfo, error)
```

Upload file part:
```go
func (b2PartInfo FilePartInfo) UploadFilePart(
	chunkNum int,
	checksum string,
	contents []byte,
) error
```

Finish large file:
```go
func (b2Auth Auth) FinishLargeFile(
	fileID string,
	checksums string,
) (LargeFile, error)
```

Cancel large file:
```go
func (b2Auth Auth) CancelLargeFile(fileID string) (bool, error)
```

___

#### Example

```go
dataSize := 10485760
chunkSize := 5242880

data := make([]byte, dataSize) // empty, only for example
var checksums []string

b2, _ := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

b2InitFile, _ := b2.StartLargeFile("mybigfile.mp4")
b2PartUploader, _ := b2.GetUploadPartURL(b2InitFile)

for i := 0; i < dataSize; i++ {
	start := i * chunkSize
	stop := i * chunkSize + chunkSize
	chunk := data[start:stop]

	h := sha1.New()
	h.Write(data)
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	checksums = append(checksums, checksum)

	// B2 chunk numbering starts at 1, not 0
	err := info.UploadFilePart(i+1, checksum, chunk)

	if err != nil {
		panic(err)
	}
}

checksumsStr := "[\"" + strings.Join(checksums, "\",\"") + "\"]"

err = b2.FinishLargeFile(b2PartUploader.FileID, checksumsStr)
if err != nil {
	panic(err)
}
```

### Download File

Downloading a file can either be done in one request (likely only
feasible for smaller files) or chunked, similar to how large files
are uploaded but in reverse.

___

#### Functions

Multi-part download:
```go
func (b2Auth Auth) PartialDownloadById(
	id string,
	begin int,
	end int,
) ([]byte, error)
```

Full download:

```go
func (b2Auth Auth) DownloadById(id string) ([]byte, error)
```

___

#### Example (single request)

```go
b2, _ := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

id := getB2FileID() // value from UploadFile or FinishLargeFile

data, err := b2.DownloadById(id)

// do something with the file data
```

#### Example (multi request)

```go
b2, _ := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

id := getB2FileID() // value from UploadFile or FinishLargeFile
fileSize := getB2FileSize() // same note as above

chunkSize := 5242880
i := 0
var output []byte

for i < fileSize {
	start := i * chunkSize
	stop := i * chunkSize + chunkSize
	data, _ := auth.PartialDownloadById(id, start, stop)
	output = append(output, data...)
}

// do something with output (full file data)
```

### Delete a File

Deleting a file requires both the file's ID, and the file's name. Both
of these are returned in the final struct when uploading a file and
should be stored somewhere if you want to delete the file later on.

___

#### Function

```go
func (b2Auth Auth) DeleteFile(b2ID string, name string) bool
```
___

#### Example

```go
b2, _ := b2.AuthorizeAccount(
	os.Getenv("B2_BUCKET_KEY_ID"),
	os.Getenv("B2_BUCKET_KEY"))

id, name := getB2FileInfo()

if b2.DeleteFile(id, name) {
	fmt.Println("File successfully deleted")
} else {
	return errors.New("failed to delete file")
}
```
