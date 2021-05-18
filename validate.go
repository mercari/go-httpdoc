package httpdoc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	"github.com/tenntenn/gpath"
)

var (
	defaultUnmarshalFunc = json.Unmarshal

	defaultAssertFunc = func(t *testing.T, expected, actual interface{}, desc string) {
		if diff := pretty.Compare(actual, expected); diff != "" {
			tFatalf(t, "%s: (-got +want): \n%s", desc, diff)
		}
	}

	defaultFatalFunc = func(t *testing.T, format string, args ...interface{}) {
		t.Fatalf(format, args...)
	}

	protoUnmarshalFunc = func(data []byte, v interface{}) error {
		unmashaler, ok := v.(proto.Unmarshaler)
		if !ok {
			return fmt.Errorf("failed to type assert to Unmashaler: %T must implement proto.Unmarshaler interface", v)
		}
		return unmashaler.Unmarshal(data)
	}
)

var tFatalf fatalFunc = defaultFatalFunc

type (
	assertFunc    func(t *testing.T, expected, actual interface{}, desc string)
	fatalFunc     func(t *testing.T, format string, args ...interface{})
	unmarshalFunc func(data []byte, v interface{}) error
)

// Validator takes test cases and checks whether recorded values are equal to the given expected values.
// If not, it fails in the given test context. If ok, it uses the result for documentation.
type Validator struct {
	record *record

	unmarshalFunc unmarshalFunc
	assertFunc    assertFunc

	requestParams  []Data
	requestHeaders []Data
	requestFields  []Data

	responseHeaders []Data
	responseFields  []Data
}

type record struct {
	requestParams  url.Values
	requestHeaders http.Header
	requestBody    []byte

	responseStatusCode int
	responseHeaders    http.Header
	responseBody       []byte
}

// TestCase is test case validator uses. Validator inspects and extract request & response value based on
// Target (e.g, when testing request params, target is parameter name. when testing response
// body, target is filed name) and asserts with Expected value.
//
// TestCase can be used like table-driven way.
//
//   validator.RequestParams(t, []httpdoc.TestCase{
//       NewTestCase("token","12345","Request token"),
//       NewTestCase("pretty","true","Pretty print response message"),
//	 })
//
type TestCase struct {
	Target      string
	Expected    interface{}
	Description string
	AssertFunc  assertFunc
}

// NewTestCase returns new TestCase.
func NewTestCase(target string, expected interface{}, description string) TestCase {
	return TestCase{Target: target, Expected: expected, Description: description}
}

func newValidator() *Validator {
	return &Validator{
		unmarshalFunc: defaultUnmarshalFunc,
		assertFunc:    defaultAssertFunc,
		record:        &record{},
	}
}

// ResponseStatusCode validates response status code is expected or not.
func (v *Validator) ResponseStatusCode(t *testing.T, expected int) {
	v.assertFunc(t, expected, v.record.responseStatusCode, "response status code")
}

// RequestParams validated request params are expected or not.
func (v *Validator) RequestParams(t *testing.T, cases []TestCase) {
	for _, tc := range cases {
		data := Data{
			Name:        tc.Target,
			Value:       tc.Expected,
			Description: tc.Description,
		}
		v.requestParams = append(v.requestParams, data)
		pickAssertFunc(&tc, v)(t, tc.Expected, v.record.requestParams.Get(tc.Target), tc.Description)
	}
}

// RequestHeaders validates request headers are expected or not.
func (v *Validator) RequestHeaders(t *testing.T, cases []TestCase) {
	for _, tc := range cases {
		data := Data{
			Name:        tc.Target,
			Value:       tc.Expected,
			Description: tc.Description,
		}
		v.requestHeaders = append(v.requestHeaders, data)

		actual := v.record.requestHeaders.Get(tc.Target)
		if actual == "" {
			h, ok := v.record.requestHeaders[tc.Target]
			if !ok || len(h) == 0 {
				tFatalf(t, "request header %q is not found", tc.Target)
				return
			}
			actual = h[0]
		}

		pickAssertFunc(&tc, v)(t, tc.Expected, actual, tc.Description)
	}
}

// ResponseHeaders validates response headers are expected or not.
func (v *Validator) ResponseHeaders(t *testing.T, cases []TestCase) {
	for _, tc := range cases {
		data := Data{
			Name:        tc.Target,
			Value:       tc.Expected,
			Description: tc.Description,
		}
		v.responseHeaders = append(v.responseHeaders, data)

		actual := v.record.responseHeaders.Get(tc.Target)
		if actual == "" {
			h, ok := v.record.responseHeaders[tc.Target]
			if !ok || len(h) == 0 {
				tFatalf(t, "request header %q is not found", tc.Target)
				return
			}
			actual = h[0]
		}
		pickAssertFunc(&tc, v)(t, tc.Expected, actual, tc.Description)
	}
}

// RequestBody validates request body's fileds are expected or not. The request body
// is unmarshaled to the given struct. To extract a filed to validate, this uses dot-seprated
// expression in TestCase.Target. For example, if you want to access `Email` value in the
// following struct use `Setting.Name` in Target.
//
//   type User struct {
//       Setting Setting
//   }
//
//   type Setting struct {
//       Email string
//   }
//
func (v *Validator) RequestBody(t *testing.T, cases []TestCase, request interface{}) {
	// Unmarshal request body into the given struct
	if err := v.unmarshalFunc(v.record.requestBody, request); err != nil {
		tFatalf(t, "Failed to unmarshal request body: %s", err)
		return
	}
	v.validateFields(t, cases, request, &v.requestFields)
}

// ResponseBody validates response body's fields are expected or not. The response body
// is unmarshaled to the given struct. To extract a filed to validate, this uses dot-seprated
// expression in TestCase.Target. For example, if you want to access `Email` value in the
// following struct use `Setting.Name` in Target.
//
//   type User struct {
//       Setting Setting
//   }
//
//   type Setting struct {
//       Email string
//   }
//
func (v *Validator) ResponseBody(t *testing.T, cases []TestCase, response interface{}) {
	// Unmarshal request body into the given struct
	if err := v.unmarshalFunc(v.record.responseBody, response); err != nil {
		tFatalf(t, "Failed to unmarshal response body: %s", err)
	}
	v.validateFields(t, cases, response, &v.responseFields)
}

func (vl *Validator) validateFields(t *testing.T, cases []TestCase, v interface{}, fields *[]Data) {
	for _, tc := range cases {
		data := Data{
			Name:        tc.Target,
			Value:       tc.Expected,
			Description: tc.Description,
		}
		*fields = append(*fields, data)
		actual, _ := gpath.At(v, tc.Target)
		pickAssertFunc(&tc, vl)(t, tc.Expected, actual, tc.Description)
	}
}

func pickAssertFunc(tc *TestCase, v *Validator) assertFunc {
	if tc.AssertFunc != nil {
		return tc.AssertFunc
	}
	return v.assertFunc
}
