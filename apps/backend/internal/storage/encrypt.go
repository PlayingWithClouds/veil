package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Encrypted blob layout: the magic, a random per-file nonce prefix, then the
// plaintext in chunks of chunkSize sealed one by one with AES-256-GCM (the
// last chunk may be shorter). Chunk i uses nonce prefix||i, so any chunk can be
// opened on its own, which keeps HTTP range requests (seeking) cheap.
const (
	encryptedMagic = "VEILENC1"
	prefixSize     = 8
	headerSize     = len(encryptedMagic) + prefixSize
	chunkSize      = 64 << 10
	tagSize        = 16
	sealedSize     = chunkSize + tagSize
)

// KeySize is the length in bytes of the key WithEncryption takes.
const KeySize = 32

// errNotEncrypted marks a blob that has no encrypted header.
var errNotEncrypted = errors.New("blob is not encrypted")

// newAEAD builds the AES-256-GCM sealer for key.
func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("encryption key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// chunkNonce is the nonce of chunk index under a file's nonce prefix.
func chunkNonce(prefix []byte, index uint64) []byte {
	nonce := make([]byte, prefixSize+4)
	copy(nonce, prefix)
	binary.BigEndian.PutUint32(nonce[prefixSize:], uint32(index))
	return nonce
}

// encryptingReader turns a plaintext stream into the encrypted layout.
type encryptingReader struct {
	source  io.Reader
	aead    cipher.AEAD
	prefix  []byte
	index   uint64
	pending []byte
	plain   []byte
	done    bool
}

// newEncryptingReader wraps source; the header is the first thing read.
func newEncryptingReader(source io.Reader, aead cipher.AEAD) (*encryptingReader, error) {
	prefix := make([]byte, prefixSize)
	if _, err := rand.Read(prefix); err != nil {
		return nil, err
	}
	header := append([]byte(encryptedMagic), prefix...)
	return &encryptingReader{source: source, aead: aead, prefix: prefix, pending: header, plain: make([]byte, chunkSize)}, nil
}

// Read returns encrypted bytes, sealing the next chunk when the buffer runs dry.
func (r *encryptingReader) Read(buffer []byte) (int, error) {
	for len(r.pending) == 0 {
		if r.done {
			return 0, io.EOF
		}
		if err := r.sealNext(); err != nil {
			return 0, err
		}
	}
	copied := copy(buffer, r.pending)
	r.pending = r.pending[copied:]
	return copied, nil
}

// sealNext reads up to one chunk of plaintext and queues its sealed form.
func (r *encryptingReader) sealNext() error {
	read, err := io.ReadFull(r.source, r.plain)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		r.done = true
		err = nil
	}
	if err != nil {
		return err
	}
	if read == 0 {
		return nil
	}
	r.pending = r.aead.Seal(nil, chunkNonce(r.prefix, r.index), r.plain[:read], nil)
	r.index++
	return nil
}

// decryptingFile reads the plaintext of an encrypted blob and supports seeking.
type decryptingFile struct {
	file       io.ReaderAt
	closer     io.Closer
	aead       cipher.AEAD
	prefix     []byte
	plainSize  int64
	position   int64
	cached     []byte
	cachedFrom int64
}

// plainSizeOf is the plaintext length of an encrypted blob of storedSize bytes.
func plainSizeOf(storedSize int64) (int64, error) {
	body := storedSize - int64(headerSize)
	if body < 0 {
		return 0, errors.New("encrypted blob shorter than its header")
	}
	full := body / sealedSize
	rest := body % sealedSize
	if rest != 0 && rest <= tagSize {
		return 0, errors.New("encrypted blob has a truncated chunk")
	}
	size := full * chunkSize
	if rest != 0 {
		size += rest - tagSize
	}
	return size, nil
}

// openEncrypted wraps file as a decrypting reader when it starts with the
// encrypted header; otherwise it returns errNotEncrypted. A nil aead means no
// key is configured, so encrypted blobs cannot be opened.
func openEncrypted(file io.ReaderAt, closer io.Closer, storedSize int64, aead cipher.AEAD) (*decryptingFile, error) {
	header := make([]byte, headerSize)
	if _, err := file.ReadAt(header, 0); err != nil || string(header[:len(encryptedMagic)]) != encryptedMagic {
		return nil, errNotEncrypted
	}
	if aead == nil {
		return nil, errors.New("blob is encrypted but no key is configured")
	}
	plainSize, err := plainSizeOf(storedSize)
	if err != nil {
		return nil, err
	}
	return &decryptingFile{
		file:       file,
		closer:     closer,
		aead:       aead,
		prefix:     header[len(encryptedMagic):],
		plainSize:  plainSize,
		cachedFrom: -1,
	}, nil
}

// Size is the plaintext length.
func (d *decryptingFile) Size() int64 {
	return d.plainSize
}

// Close closes the underlying file.
func (d *decryptingFile) Close() error {
	return d.closer.Close()
}

// Read decrypts from the current position.
func (d *decryptingFile) Read(buffer []byte) (int, error) {
	if d.position >= d.plainSize {
		return 0, io.EOF
	}
	index := d.position / chunkSize
	if err := d.load(index); err != nil {
		return 0, err
	}
	copied := copy(buffer, d.cached[d.position-index*chunkSize:])
	d.position += int64(copied)
	return copied, nil
}

// load makes chunk index the cached one, decrypting it from disk.
func (d *decryptingFile) load(index int64) error {
	if d.cachedFrom == index {
		return nil
	}
	sealed := make([]byte, sealedSize)
	read, err := d.file.ReadAt(sealed, int64(headerSize)+index*sealedSize)
	if err != nil && err != io.EOF {
		return err
	}
	plain, err := d.aead.Open(nil, chunkNonce(d.prefix, uint64(index)), sealed[:read], nil)
	if err != nil {
		return fmt.Errorf("decrypt chunk %d: %w", index, err)
	}
	d.cached = plain
	d.cachedFrom = index
	return nil
}

// Seek moves the plaintext position.
func (d *decryptingFile) Seek(offset int64, whence int) (int64, error) {
	var target int64
	switch whence {
	case io.SeekStart:
		target = offset
	case io.SeekCurrent:
		target = d.position + offset
	case io.SeekEnd:
		target = d.plainSize + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if target < 0 {
		return 0, errors.New("negative position")
	}
	d.position = target
	return target, nil
}
