package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpdoc "github.com/mercari/go-httpdoc"
)

func TestUserHandlerWithProtobuf(t *testing.T) {
	document := &httpdoc.Document{
		Name: "Example API (with protobuf)",
		ExcludeHeaders: []string{
			"Accept-Encoding",
		},
	}
	defer func() {
		if err := document.Generate("doc/protobuf.md"); err != nil {
			t.Fatalf("err: %s", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/v2/user/", httpdoc.Record(&userProtoHandler{}, document, &httpdoc.RecordOption{
		Description: "Get a user",
		ExcludeHeaders: []string{
			"User-Agent",
			"Content-Length",
		},

		WithValidate: func(validator *httpdoc.Validator) {
			validator.ResponseBody(t, []httpdoc.TestCase{
				{"Name", "Immortan Joe", "User name"},
				{"Setting.Email", "immortan@madmax.com", "User email"}},
				&UserProtoResponse{},
			)
		},

		WithProtoBuffer: &httpdoc.ProtoBufferOption{
			ResponseUnmarshaler: &UserProtoResponse{},
		},
	}))

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	req, err := http.NewRequest("GET", testServer.URL+"/v2/user/169743", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
}
