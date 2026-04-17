package resend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/realianx/terraform-provider-resend/resend"
)

func newTestClient(srv *httptest.Server) *resend.Client {
	c := resend.NewClient("test-key")
	c.BaseURL = srv.URL
	c.HTTPClient = srv.Client()
	return c
}

func TestClientDo_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header, got: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "abc"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var out struct{ ID string `json:"id"` }
	if err := c.Do(context.Background(), http.MethodGet, "/test", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "abc" {
		t.Errorf("expected abc, got %s", out.ID)
	}
}

func TestClientDo_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	err := c.Do(context.Background(), http.MethodGet, "/missing", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := resend.AsHTTPError(err)
	if !ok || httpErr.StatusCode != 404 {
		t.Errorf("expected 404 HTTPError, got: %v", err)
	}
}
