package plugins

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/dop251/goja"
)

// formDataEntriesProperty is where the FormData shim keeps its fields, for
// fetch to read back.
const formDataEntriesProperty = "__entries"

// binaryShims defines the Blob and FormData globals plugins need to upload a
// file. Blob holds bytes only (Uint8Array, ArrayBuffer or other Blob parts);
// FormData collects string and Blob fields. fetch encodes a FormData body as
// multipart/form-data itself (encodeFormData).
const binaryShims = `
(function (global) {
  function toBytes(part) {
    if (part instanceof Uint8Array) return part;
    if (part instanceof ArrayBuffer) return new Uint8Array(part);
    if (part && part.__bytes instanceof Uint8Array) return part.__bytes;
    throw new TypeError("Blob parts must be Uint8Array, ArrayBuffer or Blob");
  }
  class Blob {
    constructor(parts, options) {
      const chunks = (parts || []).map(toBytes);
      let total = 0;
      for (const chunk of chunks) total += chunk.length;
      const bytes = new Uint8Array(total);
      let offset = 0;
      for (const chunk of chunks) {
        bytes.set(chunk, offset);
        offset += chunk.length;
      }
      this.__bytes = bytes;
      this.size = total;
      this.type = options && options.type ? String(options.type) : "";
    }
  }
  class FormData {
    constructor() {
      this.__entries = [];
    }
    append(name, value, filename) {
      if (value instanceof Blob) {
        const fileName = filename === undefined ? "blob" : String(filename);
        this.__entries.push({ name: String(name), blob: value, filename: fileName });
        return;
      }
      this.__entries.push({ name: String(name), value: String(value) });
    }
  }
  global.Blob = Blob;
  global.FormData = FormData;
})(globalThis);
`

// installBinaryShims defines Blob and FormData in the runtime.
func (s *session) installBinaryShims() error {
	_, err := s.vm.RunString(binaryShims)
	return err
}

// requestBody is a fetch body ready to send: its bytes, and the Content-Type
// it dictates (the multipart boundary), empty when the caller's header stands.
type requestBody struct {
	data        []byte
	contentType string
}

// readBody converts a fetch body into bytes: a string, a Uint8Array, an
// ArrayBuffer, or a FormData (encoded as multipart/form-data).
func (s *session) readBody(body goja.Value) (requestBody, error) {
	if object, isObject := body.(*goja.Object); isObject {
		if entries := object.Get(formDataEntriesProperty); isSet(entries) {
			return s.encodeFormData(entries.ToObject(s.vm))
		}
	}
	switch exported := body.Export().(type) {
	case []byte:
		return requestBody{data: exported}, nil
	case goja.ArrayBuffer:
		return requestBody{data: exported.Bytes()}, nil
	}
	return requestBody{data: []byte(body.String())}, nil
}

// encodeFormData writes the shim's entries as a multipart/form-data body.
func (s *session) encodeFormData(entries *goja.Object) (requestBody, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	count := int(entries.Get("length").ToInteger())
	for index := 0; index < count; index++ {
		entry := entries.Get(fmt.Sprint(index)).ToObject(s.vm)
		if err := s.writeFormEntry(writer, entry); err != nil {
			return requestBody{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return requestBody{}, err
	}
	return requestBody{data: buffer.Bytes(), contentType: writer.FormDataContentType()}, nil
}

// writeFormEntry writes one FormData field: a plain value or a file part.
func (s *session) writeFormEntry(writer *multipart.Writer, entry *goja.Object) error {
	name := entry.Get("name").String()
	blob := entry.Get("blob")
	if !isSet(blob) {
		return writer.WriteField(name, entry.Get("value").String())
	}
	blobObject := blob.ToObject(s.vm)
	contentType := blobObject.Get("type").String()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header := textproto.MIMEHeader{}
	disposition := fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(name), escapeQuotes(entry.Get("filename").String()))
	header.Set("Content-Disposition", disposition)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	fileBytes, _ := blobObject.Get("__bytes").Export().([]byte)
	_, err = part.Write(fileBytes)
	return err
}

// escapeQuotes makes a string safe inside a quoted header parameter.
func escapeQuotes(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
}
