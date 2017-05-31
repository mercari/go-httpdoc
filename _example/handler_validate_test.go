package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpdoc "github.com/mercari/go-httpdoc"
)

func TestUserHandlerWithValidate(t *testing.T) {
	document := &httpdoc.Document{
		Name: "Example API (with validation)",
		ExcludeHeaders: []string{
			"Accept-Encoding",
		},
	}
	defer func() {
		if err := document.Generate("doc/validate.md"); err != nil {
			t.Fatalf("err: %s", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/v1/user", httpdoc.Record(&userHandler{}, document, &httpdoc.RecordOption{
		Description: "Create a new user",
		ExcludeHeaders: []string{
			"User-Agent",
			"Content-Length",
		},

		// WithValidate option, you can validate various http request & parameter values.
		// It checks handler gets the expected value or not and assert when it's different.
		// You can annotate what kind of value you expect (description) in each validation
		// and it will be the document.
		WithValidate: func(validator *httpdoc.Validator) {
			validator.RequestParams(t, []httpdoc.TestCase{
				{"token", "12345", "Request token"},
				{"pretty", "", "Pretty print response message"},
			})

			validator.RequestHeaders(t, []httpdoc.TestCase{
				{"X-Version", "2", "Request API version"},
			})

			validator.RequestBody(t, []httpdoc.TestCase{
				{"Name", "tcnksm", "User Name"},
				{"Email", "tcnksm@mercari.com", "User email address"},
				{"Attribute.Birthday", "1988-11-24", "User birthday YYYY-MM-DD format"}},
				&createUserRequest{},
			)

			validator.ResponseStatusCode(t, http.StatusOK)

			validator.ResponseBody(t, []httpdoc.TestCase{
				{"ID", 11241988, "User ID assigned"}},
				&createUserResponse{},
			)
		},
	}))

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	req := testNewRequest(t, testServer.URL+"/v1/user?token=12345")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("err: %s", err)
	}
}
