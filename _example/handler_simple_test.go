package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpdoc "go.mercari.io/go-httpdoc"
)

func TestUserHandlerSimple(t *testing.T) {
	document := &httpdoc.Document{
		Name:           "Example API (simple)",
		ExcludeHeaders: []string{"Content-Length"},
	}
	defer func() {
		if err := document.Generate("doc/simple.md"); err != nil {
			t.Fatalf("err: %s", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/v1/user", httpdoc.Record(&userHandler{}, document, &httpdoc.RecordOption{
		Description: "Create a new user",
	}))

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	req := testNewRequest(t, testServer.URL+"/v1/user?token=12345")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testNewRequest(t *testing.T, urlStr string) *http.Request {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", " ")
	encoder.Encode(&createUserRequest{
		Name:  "tcnksm",
		Email: "tcnksm@mercari.com",
		Attribute: attribute{
			Birthday: "1988-11-24",
		},
	})

	req, err := http.NewRequest("POST", urlStr, &buf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	req.Header.Add("X-Version", "2")
	return req
}
