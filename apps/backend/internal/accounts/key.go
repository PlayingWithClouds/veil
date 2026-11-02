package accounts

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// keyFile is where a generated key is kept when the host app gives none.
const keyFile = "account.key"

// LoadKey returns the key sessions are sealed with: derived from the host's
// encryption key when there is one (the phone's comes from the Android
// Keystore), else read from (or created in) a 0600 file in dataDir.
func LoadKey(dataDir string, hostKey []byte) ([]byte, error) {
	if len(hostKey) > 0 {
		derived := sha256.Sum256(append([]byte("veil accounts "), hostKey...))
		return derived[:], nil
	}
	path := filepath.Join(dataDir, keyFile)
	existing, err := os.ReadFile(path)
	if err == nil && len(existing) == 32 {
		return existing, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read account key: %w", err)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, fmt.Errorf("write account key: %w", err)
	}
	return key, nil
}
