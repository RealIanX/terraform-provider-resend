package resend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/realianx/terraform-provider-resend/resend"
)

func TestCreateTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/templates" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "tpl-1", "object": "template"})
	}))
	defer srv.Close()
	c := newTestClient(srv)

	id, err := c.CreateTemplate(context.Background(), resend.CreateTemplateRequest{
		Name: "Welcome", HTML: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "tpl-1" {
		t.Errorf("expected tpl-1, got %s", id)
	}
}

func TestGetTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/templates/tpl-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(resend.Template{ID: "tpl-1", Name: "Welcome", HTML: "<p>Hi</p>"})
	}))
	defer srv.Close()
	c := newTestClient(srv)

	tpl, err := c.GetTemplate(context.Background(), "tpl-1")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "Welcome" {
		t.Errorf("expected Welcome, got %s", tpl.Name)
	}
}

func TestDeleteTemplate(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/templates/tpl-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newTestClient(srv)

	if err := c.DeleteTemplate(context.Background(), "tpl-1"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("delete not called")
	}
}

func TestPublishTemplate(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/templates/tpl-1/publish" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newTestClient(srv)

	if err := c.PublishTemplate(context.Background(), "tpl-1"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("publish not called")
	}
}
