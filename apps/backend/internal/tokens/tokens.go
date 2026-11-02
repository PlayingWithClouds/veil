// Package tokens gives other clients scoped access: the phone streaming from
// the home server, scripts, agents over MCP. A token is shown once at
// creation; only its SHA-256 is stored.
//
// Scopes: read (queries), write (mutations), stream (media URLs), admin
// (managing tokens and devices; implies the others).
package tokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Scopes a token can carry.
const (
	ScopeRead   = "read"
	ScopeWrite  = "write"
	ScopeStream = "stream"
	ScopeAdmin  = "admin"
)

// knownScopes lists every scope.
var knownScopes = []string{ScopeRead, ScopeWrite, ScopeStream, ScopeAdmin}

// tokenTable is the table tokens live in.
const tokenTable = "ApiToken"

// tokenPrefix marks Veil tokens, so they're recognisable in configs.
const tokenPrefix = "veil_"

// Service creates, checks and revokes tokens.
type Service struct {
	database *db.DB
}

// New returns the service on database.
func New(database *db.DB) *Service {
	return &Service{database: database}
}

// Create makes a token and returns it with its row; the token itself is
// never stored or shown again.
func (s *Service) Create(ctx context.Context, name string, scopes []string, expiresAt *time.Time) (string, *schema.ApiToken, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, fmt.Errorf("a token needs a name")
	}
	if err := checkScopes(scopes); err != nil {
		return "", nil, err
	}
	secret, err := newSecret()
	if err != nil {
		return "", nil, err
	}
	row := db.Vars{"name": name, "scopes": scopes, "tokenHash": hash(secret)}
	if expiresAt != nil {
		row["expiresAt"] = db.FormatTime(*expiresAt)
	}
	id, err := s.database.Insert(ctx, tokenTable, row)
	if err != nil {
		return "", nil, err
	}
	token, err := s.Token(ctx, id)
	return secret, token, err
}

// checkScopes fails for an empty list or an unknown scope.
func checkScopes(scopes []string) error {
	if len(scopes) == 0 {
		return fmt.Errorf("a token needs at least one scope")
	}
	for _, scope := range scopes {
		if !slices.Contains(knownScopes, scope) {
			return fmt.Errorf("unknown scope %q (read, write, stream, admin)", scope)
		}
	}
	return nil
}

// newSecret makes a random token.
func newSecret() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return tokenPrefix + base64.RawURLEncoding.EncodeToString(random), nil
}

// hash is a token's stored form.
func hash(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// Tokens lists the tokens, newest first.
func (s *Service) Tokens(ctx context.Context) ([]schema.ApiToken, error) {
	return db.QueryAs[schema.ApiToken](ctx, s.database,
		`SELECT * FROM "ApiToken" WHERE "deletedAt" IS NULL ORDER BY "createdAt" DESC`, nil)
}

// Token loads one token's row.
func (s *Service) Token(ctx context.Context, id string) (*schema.ApiToken, error) {
	return db.Get[schema.ApiToken](ctx, s.database, tokenTable, id)
}

// Revoke deletes a token for good: a revoked token must stop working, so
// there's no trash for it.
func (s *Service) Revoke(ctx context.Context, id string) error {
	return s.database.Delete(ctx, tokenTable, id)
}

// Authenticate returns the scopes of a live token and stamps its last use;
// ok is false for an unknown, revoked or expired one.
func (s *Service) Authenticate(ctx context.Context, secret string) ([]string, bool) {
	if !strings.HasPrefix(secret, tokenPrefix) {
		return nil, false
	}
	token, err := db.QueryOneAs[schema.ApiToken](ctx, s.database,
		`SELECT * FROM "ApiToken" WHERE "tokenHash" = $hash AND "deletedAt" IS NULL`, db.Vars{"hash": hash(secret)})
	if err != nil || token == nil {
		return nil, false
	}
	if token.ExpiresAt != nil && *token.ExpiresAt < db.Now() {
		return nil, false
	}
	_, _ = s.database.Exec(ctx, `UPDATE "ApiToken" SET "lastUsedAt" = $now WHERE "id" = $id`, db.Vars{"id": token.Id, "now": db.Now()})
	return token.Scopes, true
}

// Allows reports whether scopes grant scope; admin grants everything.
func Allows(scopes []string, scope string) bool {
	return slices.Contains(scopes, scope) || slices.Contains(scopes, ScopeAdmin)
}
