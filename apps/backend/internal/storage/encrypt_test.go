package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// encryptedClient is a client that encrypts under "stream-cache/".
func encryptedClient(t *testing.T, root string, key []byte) *Client {
	t.Helper()
	client, err := New(root, "http://example.test")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := client.WithEncryption(key, "stream-cache/"); err != nil {
		t.Fatalf("with encryption: %v", err)
	}
	return client
}

// randomBytes returns n random bytes.
func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEncryptedBlobRoundTripAndAtRest(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	client := encryptedClient(t, root, randomBytes(t, KeySize))
	// Sizes around chunk boundaries, including empty.
	for _, size := range []int{0, 1, chunkSize - 1, chunkSize, chunkSize + 1, 3*chunkSize + 17} {
		plain := randomBytes(t, size)
		key := "stream-cache/mp4/blob.mp4"
		if err := client.Put(ctx, key, bytes.NewReader(plain), int64(size), "video/mp4"); err != nil {
			t.Fatalf("size %d put: %v", size, err)
		}
		stored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(key)))
		if err != nil {
			t.Fatal(err)
		}
		if size >= 64 && bytes.Contains(stored, plain[:64]) {
			t.Errorf("size %d: plaintext found on disk", size)
		}
		reader, err := client.Get(ctx, key)
		if err != nil {
			t.Fatalf("size %d get: %v", size, err)
		}
		got, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || !bytes.Equal(got, plain) {
			t.Errorf("size %d: round trip differs (err %v, got %d bytes)", size, err, len(got))
		}
	}
}

func TestEncryptedBlobServesRanges(t *testing.T) {
	ctx := context.Background()
	client := encryptedClient(t, t.TempDir(), randomBytes(t, KeySize))
	plain := randomBytes(t, 3*chunkSize+500)
	if err := client.Put(ctx, "stream-cache/mp4/a.mp4", bytes.NewReader(plain), int64(len(plain)), "video/mp4"); err != nil {
		t.Fatal(err)
	}
	// A range spanning two chunks.
	from, to := chunkSize-10, 2*chunkSize+10
	request := httptest.NewRequest(http.MethodGet, "/api/blob/stream-cache/mp4/a.mp4", nil)
	request.Header.Set("Range", "bytes="+strconv.Itoa(from)+"-"+strconv.Itoa(to))
	recorder := httptest.NewRecorder()
	client.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if !bytes.Equal(recorder.Body.Bytes(), plain[from:to+1]) {
		t.Error("range body differs from plaintext slice")
	}
	if length := recorder.Header().Get("Content-Range"); length != "bytes "+strconv.Itoa(from)+"-"+strconv.Itoa(to)+"/"+strconv.Itoa(len(plain)) {
		t.Errorf("content range = %q", length)
	}
}

func TestEncryptionOnlyCoversItsPrefixesAndKeepsOldBlobsReadable(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	plainClient, err := New(root, "http://example.test")
	if err != nil {
		t.Fatal(err)
	}
	if err := plainClient.Put(ctx, "stream-cache/mp4/old.mp4", bytes.NewReader([]byte("old plain")), 9, "video/mp4"); err != nil {
		t.Fatal(err)
	}
	client := encryptedClient(t, root, randomBytes(t, KeySize))
	if err := client.Put(ctx, "img-cache/x.jpg", bytes.NewReader([]byte("image")), 5, "image/jpeg"); err != nil {
		t.Fatal(err)
	}
	stored, _ := os.ReadFile(filepath.Join(root, "img-cache", "x.jpg"))
	if string(stored) != "image" {
		t.Error("blob outside the prefixes was encrypted")
	}
	reader, err := client.Get(ctx, "stream-cache/mp4/old.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if got, _ := io.ReadAll(reader); string(got) != "old plain" {
		t.Errorf("old plain blob read as %q", got)
	}
}

func TestEncryptedBlobNeedsItsKey(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	client := encryptedClient(t, root, randomBytes(t, KeySize))
	if err := client.Put(ctx, "stream-cache/mp4/a.mp4", bytes.NewReader([]byte("secret")), 6, "video/mp4"); err != nil {
		t.Fatal(err)
	}
	for name, other := range map[string]*Client{
		"wrong key": encryptedClient(t, root, randomBytes(t, KeySize)),
		"no key":    mustNew(t, root),
	} {
		reader, err := other.Get(ctx, "stream-cache/mp4/a.mp4")
		if err == nil {
			_, err = io.ReadAll(reader)
			reader.Close()
		}
		if err == nil {
			t.Errorf("%s: read succeeded", name)
		}
		recorder := httptest.NewRecorder()
		other.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/blob/stream-cache/mp4/a.mp4", nil))
		if recorder.Code == http.StatusOK && recorder.Body.Len() > 0 {
			t.Errorf("%s: handler served %q", name, recorder.Body.String())
		}
	}
}

func TestWithEncryptionRejectsBadKeys(t *testing.T) {
	client := mustNew(t, t.TempDir())
	if err := client.WithEncryption([]byte("short"), "stream-cache/"); err == nil {
		t.Error("short key accepted")
	}
}

// mustNew builds a plain client.
func mustNew(t *testing.T, root string) *Client {
	t.Helper()
	client, err := New(root, "http://example.test")
	if err != nil {
		t.Fatal(err)
	}
	return client
}
