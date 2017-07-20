package httpdoc

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
)

var (
	testExcludeHeaders = []string{"User-Agent", "Accept-Encoding", "Content-Length"}

	testHandler = func(w http.ResponseWriter, r *http.Request) {
		// To record, request example, request body must be read in handler
		io.Copy(ioutil.Discard, r.Body)

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("hello"))
	}

	testHandlerProto = func(w http.ResponseWriter, r *http.Request) {
		// To record, request example, request body must be read in handler
		io.Copy(ioutil.Discard, r.Body)

		response := &UserProtoResponse{
			Id:     7089,
			Name:   "tcnksm",
			Active: true,
		}
		buf, _ := response.Marshal()

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/protobuf")
		w.Write(buf)
	}
)

func TestRecord(t *testing.T) {
	cases := []struct {
		path          string
		handler       http.HandlerFunc
		recordOption  *RecordOption
		requestMethod string
		requestParam  string
		requestBody   io.Reader
		want          Entry
	}{
		{
			"/v1/hello",
			testHandler,
			nil,
			"GET",
			"?token=123456&pretty=true",
			strings.NewReader("hello"),
			Entry{
				Description: "",
				Method:      "GET",
				Path:        "/v1/hello",

				RequestParams: []Data{
					{"pretty", "true", ""},
					{"token", "123456", ""},
				},
				RequestHeaders: []Data{
					{"Accept-Encoding", "gzip", ""},
					{"Content-Length", "5", ""},
					{"User-Agent", "Go-http-client/1.1", ""},
				},
				RequestFields:  nil,
				RequestExample: "hello",

				ResponseStatusCode: http.StatusOK,
				ResponseHeaders: []Data{
					{"Content-Type", "text/plain", ""},
				},
				ResponseExample: "hello",
			},
		},

		{
			"/v1/hello",
			testHandler,
			&RecordOption{
				ExcludeHeaders: testExcludeHeaders,
			},
			"GET",
			"",
			strings.NewReader("hello"),
			Entry{
				Description: "",
				Method:      "GET",
				Path:        "/v1/hello",

				RequestParams:  []Data{},
				RequestHeaders: []Data{},
				RequestFields:  nil,
				RequestExample: "hello",

				ResponseStatusCode: http.StatusOK,
				ResponseHeaders: []Data{
					{"Content-Type", "text/plain", ""},
				},
				ResponseExample: "hello",
			},
		},

		{
			"/v1/hello",
			testHandler,
			&RecordOption{
				ExcludeHeaders: testExcludeHeaders,
				WithValidate: func(v *Validator) {
					v.RequestParams(t, []TestCase{
						{"token", "123456", "Test token"},
					})
				},
			},
			"GET",
			"?token=123456",
			strings.NewReader("hello"),
			Entry{
				Description: "",
				Method:      "GET",
				Path:        "/v1/hello",

				RequestParams: []Data{
					{"token", "123456", "Test token"},
				},
				RequestHeaders: []Data{},
				RequestFields:  nil,
				RequestExample: "hello",

				ResponseStatusCode: http.StatusOK,
				ResponseHeaders: []Data{
					{"Content-Type", "text/plain", ""},
				},
				ResponseExample: "hello",
			},
		},
	}

	for _, tc := range cases {
		document := &Document{}
		mux := http.NewServeMux()
		mux.Handle(tc.path, Record(tc.handler, document, tc.recordOption))
		testServer := httptest.NewServer(mux)

		client := http.DefaultClient
		req, err := http.NewRequest(tc.requestMethod, testServer.URL+tc.path+tc.requestParam, tc.requestBody)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := client.Do(req); err != nil {
			t.Fatal(err)
		}

		if len(document.Entries) != 1 {
			t.Fatalf("expect doc records 1 entry")
		}

		got := document.Entries[0]
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("\ngot  %#v\nwant %#v", got, tc.want)
		}
		testServer.Close()
	}
}

func TestRecord_Proto(t *testing.T) {
	cases := []struct {
		path          string
		handler       http.HandlerFunc
		recordOption  *RecordOption
		requestMethod string
		requestParam  string
		requestBody   proto.Marshaler
		want          Entry
	}{
		{
			"/v1/hello_proto",
			testHandlerProto,
			&RecordOption{
				ExcludeHeaders: testExcludeHeaders,
				WithProtoBuffer: &ProtoBufferOption{
					RequestUnmarshaler:  &UserProtoRequest{},
					ResponseUnmarshaler: &UserProtoResponse{},
				},
			},
			"GET",
			"",
			&UserProtoRequest{
				Id:   7089,
				Name: "tcnksm",
			},
			Entry{
				Description: "",
				Method:      "GET",
				Path:        "/v1/hello_proto",

				RequestParams:  []Data{},
				RequestHeaders: []Data{},
				RequestFields:  nil,
				RequestExample: `{
  "id": 7089,
  "name": "tcnksm"
}
`,

				ResponseStatusCode: http.StatusOK,
				ResponseHeaders: []Data{
					{"Content-Type", "application/protobuf", ""},
				},
				ResponseExample: `{
  "id": 7089,
  "name": "tcnksm",
  "active": true
}
`,
			},
		},
	}

	for _, tc := range cases {
		document := &Document{}
		mux := http.NewServeMux()
		mux.Handle(tc.path, Record(tc.handler, document, tc.recordOption))
		testServer := httptest.NewServer(mux)

		client := http.DefaultClient
		buf, err := tc.requestBody.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(tc.requestMethod, testServer.URL+tc.path+tc.requestParam, bytes.NewReader(buf))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := client.Do(req); err != nil {
			t.Fatal(err)
		}

		if len(document.Entries) != 1 {
			t.Fatalf("expect doc records 1 entry")
		}

		got := document.Entries[0]
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("\ngot  %#v\nwant %#v", got, tc.want)
		}
		testServer.Close()
	}
}

func TestConvertHeaders(t *testing.T) {
	input := map[string][]string{
		"Content-Type":  []string{"application/json"},
		"X-API-Version": []string{"1.1.2"},
	}

	got := convertHeaders(input)
	want := []Data{
		{
			Name:        "Content-Type",
			Value:       "application/json",
			Description: "",
		},
		{
			Name:        "X-API-Version",
			Value:       "1.1.2",
			Description: "",
		},
	}

	sort.Sort(byName(got))

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestMergeData(t *testing.T) {
	a := []Data{
		{Name: "1"},
		{Name: "2"},
		{Name: "3", Description: "this is 3"},
		{Name: "4", Description: "this is 4"},
	}

	b := []Data{
		{Name: "3"},
		{Name: "4"},
		{Name: "5"},
	}

	got := mergeData(a, b)
	want := []Data{
		{Name: "1"},
		{Name: "2"},
		{Name: "3", Description: "this is 3"},
		{Name: "4", Description: "this is 4"},
		{Name: "5"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestExcludeData(t *testing.T) {
	input := []Data{
		{Name: "1"},
		{Name: "2"},
		{Name: "3"},
		{Name: "4"},
		{Name: "5"},
	}

	got := excludeData(input, []string{"1", "2"}, []string{"3"}, []string{"4"})
	want := []Data{
		{
			Name: "5",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestResponseWriter_Write(t *testing.T) {
}

func TestResponseWriter_WriteHeader(t *testing.T) {
}
