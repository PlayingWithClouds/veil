package files

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// oshashChunkBytes is how much of the head and the tail oshash reads.
	oshashChunkBytes = 64 * 1024
	// md5MaxBytes bounds the files that get an md5: it reads the whole file.
	md5MaxBytes = 100 * 1024 * 1024
)

// hashValue is one computed fingerprint.
type hashValue struct {
	algorithm schema.FingerprintAlgorithm
	value     string
}

// computeHashes fingerprints content of size bytes: always oshash, md5 too
// for small files. Empty files have none.
func computeHashes(content io.ReadSeeker, size int64) ([]hashValue, error) {
	if size <= 0 {
		return nil, nil
	}
	oshash, err := computeOshash(content, size)
	if err != nil {
		return nil, err
	}
	hashes := []hashValue{{algorithm: schema.FingerprintAlgorithmOshash, value: oshash}}
	if size > md5MaxBytes {
		return hashes, nil
	}
	md5Value, err := computeMD5(content)
	if err != nil {
		return nil, err
	}
	return append(hashes, hashValue{algorithm: schema.FingerprintAlgorithmMd5, value: md5Value}), nil
}

// computeOshash is the OpenSubtitles hash (as stash and StashDB use it):
// the file size plus the little-endian 64-bit words of the first and last
// 64 KiB, as 16 hex digits. Smaller files use the whole file for both.
func computeOshash(content io.ReadSeeker, size int64) (string, error) {
	chunkBytes := int64(oshashChunkBytes)
	if size < chunkBytes {
		chunkBytes = size
	}
	head, err := readAt(content, 0, chunkBytes)
	if err != nil {
		return "", err
	}
	tail, err := readAt(content, size-chunkBytes, chunkBytes)
	if err != nil {
		return "", err
	}
	sum := uint64(size) + sumWords(head) + sumWords(tail)
	return fmt.Sprintf("%016x", sum), nil
}

// readAt reads length bytes of content from offset.
func readAt(content io.ReadSeeker, offset, length int64) ([]byte, error) {
	if _, err := content.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	buffer := make([]byte, length)
	if _, err := io.ReadFull(content, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

// sumWords adds up the whole little-endian 64-bit words of data, wrapping;
// trailing bytes that don't fill a word are ignored.
func sumWords(data []byte) uint64 {
	var sum uint64
	for offset := 0; offset+8 <= len(data); offset += 8 {
		sum += binary.LittleEndian.Uint64(data[offset : offset+8])
	}
	return sum
}

// computeMD5 is the hex md5 of all of content.
func computeMD5(content io.ReadSeeker) (string, error) {
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, content); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
