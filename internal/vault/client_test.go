package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestNewClient_MissingAddress(t *testing.T) {
	_, err := NewClient(Config{Token: "tok"})
	if err == nil {
		t.Fatal("expected error for missing address, got nil")
	}
}

func TestNewClient_MissingToken(t *testing.T) {
	_, err := NewClient(Config{Address: "http://localhost:8200"})
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

func TestNewClient_Namespace(t *testing.T) {
	client, err := NewClient(Config{
		Address:   "http://localhost:8200",
		Token:     "test-token",
		Namespace: "team-a",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Namespace() != "team-a" {
		t.Errorf("expected namespace %q, got %q", "team-a", client.Namespace())
	}
}

func TestReadSecret_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/data/myapp/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"DB_PASS": "s3cr3t",
				},
			},
		})
	})

	client, err := NewClient(Config{Address: srv.URL, Token: "test"})
	if err != nil {
		t.Fatalf("client creation failed: %v", err)
	}

	data, err := client.ReadSecret("myapp/config")
	if err != nil {
		t.Fatalf("ReadSecret failed: %v", err)
	}
	if data["DB_PASS"] != "s3cr3t" {
		t.Errorf("expected DB_PASS=s3cr3t, got %v", data["DB_PASS"])
	}
}

func TestKvV2DataPath(t *testing.T) {
	got := kvV2DataPath("myapp/config")
	want := "secret/data/myapp/config"
	if got != want {
		t.Errorf("kvV2DataPath: want %q, got %q", want, got)
	}
}
