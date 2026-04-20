package providers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDevopsListValuesUsesBasicAuthAndRedirectSuppressHeaders(t *testing.T) {
	t.Helper()

	var sawAuthorization string
	var sawAccept string
	var sawRedirectSuppress string
	var sawPassThrough string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuthorization = r.Header.Get("Authorization")
		sawAccept = r.Header.Get("Accept")
		sawRedirectSuppress = r.Header.Get("X-TFS-FedAuthRedirect")
		sawPassThrough = r.Header.Get("X-VSS-ForceMsaPassThrough")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"name":"Azurefox Proof Lab"}]}`))
	}))
	defer server.Close()

	values, err := devopsListValuesWithClient(context.Background(), "abc123", server.URL, server.Client())
	if err != nil {
		t.Fatalf("devopsListValuesWithClient() error = %v, want nil", err)
	}
	if len(values) != 1 || stringMapValue(values[0], "name") != "Azurefox Proof Lab" {
		t.Fatalf("devopsListValuesWithClient() values = %#v, want one parsed project row", values)
	}

	wantAuthorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(":abc123"))
	if sawAuthorization != wantAuthorization {
		t.Fatalf("Authorization header = %q, want %q", sawAuthorization, wantAuthorization)
	}
	if sawAccept != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", sawAccept)
	}
	if sawRedirectSuppress != "Suppress" {
		t.Fatalf("X-TFS-FedAuthRedirect = %q, want Suppress", sawRedirectSuppress)
	}
	if sawPassThrough != "true" {
		t.Fatalf("X-VSS-ForceMsaPassThrough = %q, want true", sawPassThrough)
	}
}

func TestDevopsListValuesSurfacesNonJSONResponseTruthfully(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>sign in required</body></html>"))
	}))
	defer server.Close()

	_, err := devopsListValuesWithClient(context.Background(), "abc123", server.URL, server.Client())
	if err == nil {
		t.Fatal("devopsListValuesWithClient() error = nil, want non-JSON response error")
	}
	if !strings.Contains(err.Error(), `received non-JSON Azure DevOps response`) {
		t.Fatalf("error = %q, want non-JSON Azure DevOps response guidance", err)
	}
	if !strings.Contains(err.Error(), `text/html; charset=utf-8`) {
		t.Fatalf("error = %q, want content-type included", err)
	}
	if !strings.Contains(err.Error(), `sign in required`) {
		t.Fatalf("error = %q, want body snippet included", err)
	}
}
