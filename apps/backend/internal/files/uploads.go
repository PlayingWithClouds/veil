package files

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// UploadRoute is where clients upload what the backend can't make itself:
// subtitle tracks, motion scripts, trickplay sheets, exported clips,
// frames for thumbnails.
const UploadRoute = "/api/upload"

// uploadPrefix is the blob prefix of uploads; it's outside the cache, so
// uploads are never evicted.
const uploadPrefix = "uploads/"

// maxUploadBytes bounds one upload (an exported clip).
const maxUploadBytes = 4 << 30

// Uploaded is the answer to an upload: the asset URI to store and the URL
// to fetch it, and the file row when it was recorded for a record.
type Uploaded struct {
	URI    string `json:"uri"`
	URL    string `json:"url"`
	FileID string `json:"fileId,omitempty"`
}

// UploadHandler serves POST /api/upload?name=<file name>: the request body
// is the file (Content-Type its type). With subjectKind and subjectId the
// file is also recorded as a MediaFile of that record (an exported clip's
// file, a recording), whose id the answer carries.
func (s *Service) UploadHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "POST the file as the body", http.StatusMethodNotAllowed)
			return
		}
		uploaded, err := s.storeUpload(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(uploaded)
	})
}

// storeUpload writes the body to the blob store and records the file when
// the request names a record.
func (s *Service) storeUpload(request *http.Request) (*Uploaded, error) {
	ctx := request.Context()
	name := request.URL.Query().Get("name")
	key := uploadPrefix + db.NewID() + uploadExtension(name, request.Header.Get("Content-Type"))
	body := http.MaxBytesReader(nil, request.Body, maxUploadBytes)
	counted := &countingReader{reader: body}
	if err := s.config.Store.Put(ctx, key, counted, request.ContentLength, request.Header.Get("Content-Type")); err != nil {
		return nil, err
	}
	uploaded := &Uploaded{URI: blobScheme + key, URL: s.config.Store.DirectURL(key)}
	kind := schema.RecordKind(request.URL.Query().Get("subjectKind"))
	id := request.URL.Query().Get("subjectId")
	if kind == "" || id == "" {
		return uploaded, nil
	}
	fileID, err := s.recordUpload(request, kind, id, key, counted.read)
	if err != nil {
		return nil, err
	}
	uploaded.FileID = fileID
	return uploaded, nil
}

// recordUpload records an uploaded blob as a file of the record.
func (s *Service) recordUpload(request *http.Request, kind schema.RecordKind, id, key string, size int64) (string, error) {
	location, err := s.libraryLocation(request)
	if err != nil {
		return "", err
	}
	mimeType := request.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return s.database.Insert(request.Context(), tableMediaFile, db.Vars{
		"subjectKind": kind, "subjectId": id, "location": location.Id, "path": key,
		"sizeBytes": size, "mimeType": mimeType, "missing": false,
	})
}

// libraryLocation is this device's library location in the blob store.
func (s *Service) libraryLocation(request *http.Request) (*schema.StorageLocation, error) {
	locations, err := s.config.Devices.Locations(request.Context(), "")
	if err != nil {
		return nil, err
	}
	for index := range locations {
		if locations[index].Kind == schema.StorageKindLibrary && s.isBlobStore(locations[index]) {
			return &locations[index], nil
		}
	}
	return nil, fmt.Errorf("this device has no library location for the blob store")
}

// blobScheme marks asset URIs of stored blobs.
const blobScheme = "blob://"

// AssetURL is where to fetch an asset: the blob store's URL for a blob://
// URI, else the URI itself.
func (s *Service) AssetURL(uri string) string {
	key, isBlob := strings.CutPrefix(uri, blobScheme)
	if !isBlob {
		return uri
	}
	return s.config.Store.DirectURL(key)
}

// uploadExtension is the extension a stored upload keeps: the file name's,
// else one for its type.
func uploadExtension(name, contentType string) string {
	extension := strings.ToLower(path.Ext(name))
	if extension != "" && len(extension) <= 10 {
		return extension
	}
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		return ""
	}
	return extensions[0]
}

// countingReader counts the bytes read through it.
type countingReader struct {
	reader io.Reader
	read   int64
}

// Read reads from the wrapped reader, counting the bytes.
func (c *countingReader) Read(buffer []byte) (int, error) {
	count, err := c.reader.Read(buffer)
	c.read += int64(count)
	return count, err
}
