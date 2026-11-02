package tokens

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// contextKey keys the request's access in a context.
type contextKey struct{}

// Access is what a request may do: everything for a trusted caller, else
// what its token's scopes grant.
type Access struct {
	Trusted bool
	Scopes  []string
}

// Allows reports whether the access grants scope.
func (a Access) Allows(scope string) bool {
	return a.Trusted || Allows(a.Scopes, scope)
}

// FromContext returns the request's access; requests that never passed the
// middleware (tests, background work) are trusted.
func FromContext(ctx context.Context) Access {
	access, found := ctx.Value(contextKey{}).(Access)
	if !found {
		return Access{Trusted: true}
	}
	return access
}

// WithAccess returns ctx carrying access.
func WithAccess(ctx context.Context, access Access) context.Context {
	return context.WithValue(ctx, contextKey{}, access)
}

// mediaPrefixes are the routes that serve media and need the stream scope.
var mediaPrefixes = []string{"/api/stream/", "/api/blob/", "/api/file/", "/api/img"}

// Middleware checks tokens when required() is true. Loopback callers (the
// web server and the phone app on the same machine) are always trusted;
// others need a token as "Authorization: Bearer …" or, for media URLs a
// player can't add headers to, "?token=…". Media routes need the stream
// scope; GraphQL's scopes are checked per operation (see Access). The health
// check stays open.
func (s *Service) Middleware(required func() bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !required() || isLoopback(request) || request.URL.Path == "/health" || request.Method == http.MethodOptions {
			next.ServeHTTP(writer, request.WithContext(WithAccess(request.Context(), Access{Trusted: true})))
			return
		}
		scopes, ok := s.Authenticate(request.Context(), tokenOf(request))
		if !ok {
			http.Error(writer, "a valid API token is required", http.StatusUnauthorized)
			return
		}
		if isMedia(request.URL.Path) && !Allows(scopes, ScopeStream) {
			http.Error(writer, "this token can't stream", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request.WithContext(WithAccess(request.Context(), Access{Scopes: scopes})))
	})
}

// isLoopback reports whether the request comes from this machine.
func isLoopback(request *http.Request) bool {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

// tokenOf reads the request's token from the Authorization header or the
// token query parameter.
func tokenOf(request *http.Request) string {
	header := request.Header.Get("Authorization")
	if bearer, found := strings.CutPrefix(header, "Bearer "); found {
		return strings.TrimSpace(bearer)
	}
	return request.URL.Query().Get("token")
}

// isMedia reports whether a path serves media.
func isMedia(path string) bool {
	for _, prefix := range mediaPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
