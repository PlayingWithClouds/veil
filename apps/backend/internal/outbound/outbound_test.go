package outbound

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// TestHandshakeOffersNoPostQuantumKeyShare checks that the shared client
// never offers X25519MLKEM768, which Pornhub rejects.
func TestHandshakeOffersNoPostQuantumKeyShare(t *testing.T) {
	var offered []tls.CurveID
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	server.TLS = &tls.Config{
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			offered = hello.SupportedCurves
			return nil, nil
		},
	}
	server.StartTLS()
	defer server.Close()

	transport := NewTransport()
	transport.TLSClientConfig.RootCAs = server.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs
	response, err := (&http.Client{Transport: transport}).Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	if len(offered) == 0 {
		t.Fatal("no curves offered")
	}
	if slices.Contains(offered, tls.X25519MLKEM768) {
		t.Fatalf("offered curves %v include X25519MLKEM768", offered)
	}
}
