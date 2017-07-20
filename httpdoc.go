package httpdoc

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/golang/protobuf/proto"
)

const (
	// EnvHTTPDoc is the environmental variable that determines if Generate func generates documentation
	// to the given file or not. By default, it does not generate. If this variable is not empty, then it does.
	EnvHTTPDoc = "HTTPDOC"
)

// Document stores recorded results by Record middleware.
type Document struct {
	// Name is API documentation name.
	Name string

	// ExcludeHeaders is list of headers to exclude from documentation.
	// For example, you may do not need `Content-Length` header. This is applied all entries (endpoints).
	// If you want to exclude header only in specific endpoint, then use `RecordOption.ExcludeHeaders`.
	ExcludeHeaders []string

	// Entries stores all recorded results by Record middleware. Normally, you don't need to modify this.
	// This is exported just for templating.
	Entries []Entry

	// tmpl is template file to use. Currently this is only static/tmpl/doc.md.tmpl
	tmpl string

	logger *log.Logger
}

// Entry is recorded results by Record middleware. Normally, you don't need to modify this.
// All fields are exported just for templating.
type Entry struct {
	// Description is description of endpoint.
	Description string

	// Method is HTTP method.
	Method string

	// Path is request path.
	Path string

	RequestParams  []Data
	RequestHeaders []Data
	RequestFields  []Data

	// RequestExample is request body example. If you use plain text for json for response body
	// it uses it here without modification. If you use protocol buffer format for your request body
	// it unmarshals it in the given struct and encodes it into json format.
	RequestExample string

	ResponseStatusCode int
	ResponseHeaders    []Data
	ResponseFields     []Data

	// ResponseExample is response body example. If you use plain text for json for response body
	// it uses it here without modification. If you use protocol buffer format for your response body
	// it unmarshals it in the given struct and encodes it into json format.
	ResponseExample string
}

// RecordOption is option for Record middleware.
type RecordOption struct {
	// Description is description of endpoint. This is used for Entry.Description.
	Description string

	// ExcludeHeaders is list of headers to exclude from documentation.
	// This is applied only one entry (endpoint). If you want to exclude header in all endpoints
	// use `Document.ExcludeHeaders`.
	ExcludeHeaders []string

	// WithValidate option, you can validate various http request & response parameter values.
	// It inspects values which handler receives and checks it's expected or not.
	// If not it asserts and fails the test. If ok, uses it for documentaion entry.
	//
	// Not only validate, you can add an annotation to each values (e.g., what does the header
	// means?) and it's used for documentation.
	//
	// See more usage in Validator methods.
	WithValidate func(*Validator)

	// WithProtoBuffer option is used for protocol buffer request & response.
	WithProtoBuffer *ProtoBufferOption
}

// ProtoBufferOption is option for protocol buffer.
type ProtoBufferOption struct {
	// RequestUnmarshaler is used to unmarshal protocol buffer encoded request body.
	// This is used for generating human readable request example (json format).
	RequestUnmarshaler proto.Unmarshaler

	// ResponseUnmarshaler is used to unmarshal protocol buffer encoded response body.
	// This is used for generating human readable response example (json format).
	ResponseUnmarshaler proto.Unmarshaler
}

// Data represents a request or response parameter value. Normally, you don't need to modify this.
// All fields are exported just for templating.
type Data struct {
	// Name is header or params, field name.
	Name string

	// Value is actual value handler receives.
	Value interface{}

	// Description is description for this data. You can provide this via a validator.
	Description string
}

type byName []Data

func (n byName) Len() int           { return len(n) }
func (n byName) Less(i, j int) bool { return n[i].Name < n[j].Name }
func (n byName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

// Record is a http middleware. It records all request & response values which the given http handler
// receives & response and save it in the given Document.
func Record(next http.Handler, document *Document, opt *RecordOption) http.Handler {

	// If not option is provided, initialize it.
	if opt == nil {
		opt = &RecordOption{}
	}

	if document.logger == nil {
		document.logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new responseWriter it captures status code and response body.
		rw := responseWriter{
			ResponseWriter: w,
		}

		// Create a tee reader and stores request body.
		// Because of this, handler must read request body to record.
		var requestBody bytes.Buffer
		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &requestBody))

		next.ServeHTTP(&rw, r)

		// If protobuffer option is provided, use protoUnmarshalFunc for
		// validator, by default, use json unmashal func.
		unmarshalFunc := defaultUnmarshalFunc
		if opt.WithProtoBuffer != nil {
			unmarshalFunc = protoUnmarshalFunc
		}

		validator := &Validator{
			record: &record{
				requestParams:  r.URL.Query(),
				requestHeaders: r.Header,
				requestBody:    requestBody.Bytes(),

				responseStatusCode: rw.statusCode,
				responseHeaders:    rw.Header(),
				responseBody:       rw.responseBody,
			},
			unmarshalFunc: unmarshalFunc,
			assertFunc:    defaultAssertFunc,
		}

		if opt.WithValidate != nil {
			opt.WithValidate(validator)
		}

		excludeHeaders := append(opt.ExcludeHeaders, document.ExcludeHeaders...)

		requestParams := mergeData(validator.requestParams, convertHeaders(r.URL.Query()))

		requestHeaders := mergeData(validator.requestHeaders, convertHeaders(r.Header))
		requestHeaders = excludeData(requestHeaders, excludeHeaders)

		responseHeaders := mergeData(validator.responseHeaders, convertHeaders(rw.Header()))
		responseHeaders = excludeData(responseHeaders, excludeHeaders)

		requestExample := requestBody.String()
		responseExample := string(rw.responseBody)
		if opt.WithProtoBuffer != nil {
			// FIXME(tcnksm): Want to use jsonpb but sometimes panic happens while marshalling....
			if unmarshaler := opt.WithProtoBuffer.RequestUnmarshaler; unmarshaler != nil {
				unmarshaler.Unmarshal(requestBody.Bytes())

				var buf bytes.Buffer
				encoder := json.NewEncoder(&buf)
				encoder.SetIndent("", "  ")
				encoder.Encode(unmarshaler)

				requestExample = buf.String()
			}

			if unmarshaler := opt.WithProtoBuffer.ResponseUnmarshaler; unmarshaler != nil {
				unmarshaler.Unmarshal(rw.responseBody)

				var buf bytes.Buffer
				encoder := json.NewEncoder(&buf)
				encoder.SetIndent("", "  ")
				encoder.Encode(unmarshaler)

				responseExample = buf.String()
			}
		}

		entry := Entry{
			Description: opt.Description,

			Method: r.Method,
			Path:   r.URL.Path,

			RequestHeaders: requestHeaders,
			RequestParams:  requestParams,
			RequestFields:  validator.requestFields,
			RequestExample: requestExample,

			ResponseStatusCode: rw.statusCode,
			ResponseHeaders:    responseHeaders,
			ResponseFields:     validator.responseFields,
			ResponseExample:    responseExample,
		}
		entry.format()
		document.Entries = append(document.Entries, entry)
	})
}

// format sorts entry data to prevent results updated everytime.
func (e *Entry) format() error {
	sort.Sort(byName(e.RequestHeaders))
	sort.Sort(byName(e.RequestParams))

	return nil
}

type responseWriter struct {
	statusCode   int
	responseBody []byte

	http.ResponseWriter
}

func (w *responseWriter) Write(buf []byte) (int, error) {
	w.responseBody = buf
	return w.ResponseWriter.Write(buf)
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// convertHeaders convert HTTP header to httpdoc description format.
func convertHeaders(headers map[string][]string) []Data {
	d := make([]Data, 0, len(headers))
	for k, v := range headers {
		data := Data{
			Name:  k,
			Value: v[0],
		}
		d = append(d, data)
	}
	return d
}

// mergeData merges 2 Data slice into 1 slice without duplication.
// If duplicated, item in 1st slice is used.
func mergeData(a, b []Data) []Data {
	newData := make([]Data, len(a))
	copy(newData, a)
	for _, d1 := range b {
		var contain bool
		for _, d2 := range a {
			if d1.Name == d2.Name {
				contain = true
			}
		}
		if !contain {
			newData = append(newData, d1)
		}
	}
	return newData
}

// excludeData excludes data which is given.
func excludeData(target []Data, excludes ...[]string) []Data {
	newData := make([]Data, 0, len(target))
	for _, d := range target {
		var contain bool
		for _, exclude := range excludes {
			for _, name := range exclude {
				if d.Name == name {
					contain = true
				}
			}
		}

		if !contain {
			newData = append(newData, d)
		}
	}
	return newData
}
