package httpdoc

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
)

type User struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Active     bool           `json:"active"`
	Setting    *Setting       `json:"setting"`
	Permission []string       `json:"permission"`
	Preference map[string]int `json:"preference"`
}

type Setting struct {
	Email string `json:"email"`
	SNS   SNS    `json:"sns"`
}

type SNS struct {
	Twitter  string `json:"twitter"`
	Facebook string `json:"facebook"`
}

// testAssertWithCount returns assertFunc it counts failed test instead of fail.
func testAssertWithCount(fails *int) assertFunc {
	return func(t *testing.T, expected, actual interface{}, desc string) {
		if !reflect.DeepEqual(expected, actual) {
			*fails++
		}
	}
}

func fprintFatalFunc(w io.Writer) fatalFunc {
	return func(t *testing.T, format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}
}

func TestValidator_ResponseStatusCode(t *testing.T) {
	validator := newValidator()

	validator.record.responseStatusCode = 200
	validator.ResponseStatusCode(t, 200)

	validator.record.responseStatusCode = 500
	validator.ResponseStatusCode(t, 500)

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.record.responseStatusCode = 500
	validator.ResponseStatusCode(t, 200)
	if want := 1; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}
}

func TestValidator_RequestParams(t *testing.T) {
	validator := newValidator()
	validator.record.requestParams = map[string][]string{
		"token":  []string{"12345"},
		"pretty": []string{"true"},
		"year":   []string{strconv.Itoa(time.Now().Year())},
	}
	thisYearcalledAssertFunc := false
	validator.RequestParams(t, []TestCase{
		{"token", "12345", "", nil},
		{"pretty", "true", "", nil},
		{"year", "thisyear", "", func(t *testing.T, expected, actual interface{}, desc string) {
			if expected != "thisyear" {
				t.Fatal("expected is not thisyear")
			}
			thisYearcalledAssertFunc = true
		}},
	})

	if thisYearcalledAssertFunc == false {
		t.Fatal("thisYear AssertFunc should be called.")
	}

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestParams(t, []TestCase{
		{"token", "8976", "", nil},
		{"pretty", "", "", nil},
		{"id", "u8988", "", nil},
	})
	if want := 3; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}
}

func TestValidator_RequestHeaders(t *testing.T) {
	validator := newValidator()
	validator.record.requestHeaders = map[string][]string{
		"User-Agent":    []string{"Googlebot/2.1"},
		"Content-Type":  []string{"application/json"},
		"X-API-Version": []string{"1.1.2"},
	}
	validator.RequestHeaders(t, []TestCase{
		{"User-Agent", "Googlebot/2.1", "", nil},
		{"Content-Type", "application/json", "", nil},
		{"X-API-Version", "1.1.2", "", nil},
	})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestHeaders(t, []TestCase{
		{"User-Agent", []string{"curl"}, "", nil},
		{"Content-Type", []string{"application/protobuf"}, "", nil},
		{"X-API-Version", []string{"3.0"}, "", nil},
	})
	if want := 3; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}

	var buf bytes.Buffer
	tFatalf = fprintFatalFunc(&buf)
	validator.RequestHeaders(t, []TestCase{
		{"Not-Found", []string{""}, "", nil},
	})

	if got, want := buf.String(), "not found"; !strings.Contains(got, want) {
		t.Fatalf("expect %q to contain %q", got, want)
	}
}

func TestValidator_ResponseHeaders(t *testing.T) {
	validator := newValidator()
	rand.Seed(time.Now().UnixNano())
	length := 0
	for {
		length = rand.Intn(100000)
		if length > 0 {
			break
		}
	}

	validator.record.responseHeaders = map[string][]string{
		"Content-Type":   []string{"application/json"},
		"X-API-Version":  []string{"1.1.2"},
		"Content-Length": []string{strconv.Itoa(length)},
	}
	contentLengthCalledAssertFunc := false
	validator.ResponseHeaders(t, []TestCase{
		{"Content-Type", "application/json", "", nil},
		{"X-API-Version", "1.1.2", "", nil},
		{"Content-Length", []string{"content length"}, "length is change every time", func(t *testing.T, expected, actual interface{}, desc string) {
			contentLength, err := strconv.Atoi(actual.(string))
			if err != nil {
				t.Fatal("actual is not number")
			}
			if contentLength <= 0 {
				t.Fatal("actual must be greater than 0")
			}
			contentLengthCalledAssertFunc = true
		}},
	})

	if contentLengthCalledAssertFunc == false {
		t.Fatal("content length AssertFunc should be called.")
	}

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseHeaders(t, []TestCase{
		{"Content-Type", []string{"application/protobuf"}, "", nil},
	})
	if want := 1; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}

	var buf bytes.Buffer
	tFatalf = fprintFatalFunc(&buf)
	validator.ResponseHeaders(t, []TestCase{
		{"Not-Found", []string{""}, "", nil},
	})

	if got, want := buf.String(), "not found"; !strings.Contains(got, want) {
		t.Fatalf("expect %q to contain %q", got, want)
	}
}

func TestValidator_RequestBody(t *testing.T) {
	validator := newValidator()
	validator.record.requestBody = []byte(`{
  "id": 910,
  "setting": {
    "email": "taichi@mercari.com"
  }
}
`)
	validator.RequestBody(t, []TestCase{
		{"ID", 910, "", nil},
		{"Setting.Email", "taichi@mercari.com", "", nil},
	}, &User{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestBody(t, []TestCase{
		{"ID", 123, "", nil},
		{"Active", true, "", nil},
		{"Setting.Email", "deeeet@gmail.com", "", nil},
	}, &User{})

	if want := 3; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}

	var buf bytes.Buffer
	tFatalf = fprintFatalFunc(&buf)
	validator.RequestBody(t, []TestCase{}, struct{}{})
	if got, want := buf.String(), "Failed to unmarshal request"; !strings.Contains(got, want) {
		t.Fatalf("expect %q to contain %q", got, want)
	}
}

func TestValidator_ResponseBody(t *testing.T) {
	validator := newValidator()
	validator.record.responseBody = []byte(`{
  "id": 789,
  "active": false,
  "setting": {
    "email": "tcnksm@mercari.com"
  },
  "permission": ["write","read"],
  "preference": {
    "email": 0
  }
}
`)
	custommailCalledAssertFunc := false
	validator.ResponseBody(t, []TestCase{
		{"ID", 789, "", nil},
		{"Active", false, "", nil},
		{"Setting.Email", "tcnksm@mercari.com", "", nil},
		{"Setting.Email", "custommail", "", func(t *testing.T, expected, actual interface{}, desc string) {
			if expected != "custommail" {
				t.Fatal("Setting.Email is not custommail")
			}
			custommailCalledAssertFunc = true
		}},
		{"Permission[1]", "read", "", nil},
		{`Preference["email"]`, 0, "", nil},
	}, &User{})

	if custommailCalledAssertFunc == false {
		t.Fatal("custom mail AssertFunc should be called.")
	}

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseBody(t, []TestCase{
		{"ID", 123, "", nil},
		{"Active", true, "", nil},
		{"Setting.Email", "deeeet@gmail.com", "", nil},
		{"Permission[1]", "write", "", nil},
		{`Preference["email"]`, 1, "", nil},
	}, &User{})

	if want := 5; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}

	var buf bytes.Buffer
	tFatalf = fprintFatalFunc(&buf)
	validator.ResponseBody(t, []TestCase{}, struct{}{})
	if got, want := buf.String(), "Failed to unmarshal response"; !strings.Contains(got, want) {
		t.Fatalf("expect %q to contain %q", got, want)
	}
}

func TestValidateFields(t *testing.T) {
	testUser := &User{
		ID:     12345,
		Name:   "tcnksm",
		Active: true,
		Setting: &Setting{
			Email: "tcnksm@example.com",
			SNS: SNS{
				Twitter: "@deeeet",
			},
		},
		Permission: []string{"write", "read"},
		Preference: map[string]int{
			"email": 0,
			"push":  1,
		},
	}

	activeCalledAssertFunc := false
	validator := newValidator()
	validator.validateFields(t, []TestCase{
		{"ID", 12345, "", nil},
		{"Name", "tcnksm", "", nil},
		{"Active", true, "", nil},
		{"Active", "customactive", "", func(t *testing.T, expected, actual interface{}, desc string) {
			if expected != "customactive" {
				t.Fatal("Acitve is not customactive")
			}
			activeCalledAssertFunc = true
		}},
		{"Setting.Email", "tcnksm@example.com", "", nil},
		{"Setting.SNS.Twitter", "@deeeet", "", nil},
		{"Permission[0]", "write", "", nil},
		{`Preference["email"]`, 0, "", nil},
	}, testUser, &[]Data{})

	if activeCalledAssertFunc == false {
		t.Fatal("active AssertFunc should be called.")
	}
}

func TestValidator_RequestBody_Proto(t *testing.T) {
	buf, err := proto.Marshal(&UserProtoRequest{
		Id:   12345,
		Name: "tcnksm",
	})
	if err != nil {
		t.Fatal(err)
	}

	validator := newValidator()
	validator.record.requestBody = buf
	validator.unmarshalFunc = protoUnmarshalFunc
	customIDCalledAssertFunc := false
	validator.RequestBody(t, []TestCase{
		{"Id", int32(12345), "", nil},
		{"Name", "tcnksm", "", nil},
		{"Id", "customid", "custom assert func test", func(t *testing.T, expected, actual interface{}, desc string) {
			if expected != "customid" {
				t.Fatal("expected is not customid")
			}
			customIDCalledAssertFunc = true
		}},
	}, &UserProtoRequest{})

	if customIDCalledAssertFunc == false {
		t.Fatal("custom id AssertFunc should be called.")
	}

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestBody(t, []TestCase{
		{"Id", 123, "", nil},
	}, &UserProtoRequest{})

	if want := 1; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}
}

func TestValidator_ResponseBody_Proto(t *testing.T) {
	buf, err := proto.Marshal(&UserProtoResponse{
		Id: 667854,
		Setting: &UserProtoResponse_Setting{
			Email: "httpdoc@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	validator := newValidator()
	validator.unmarshalFunc = protoUnmarshalFunc
	validator.record.responseBody = buf
	validator.ResponseBody(t, []TestCase{
		{"Id", int32(667854), "", nil},
		{"Setting.Email", "httpdoc@example.com", "", nil},
	}, &UserProtoResponse{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseBody(t, []TestCase{
		{"Id", 123, "", nil},
		{"Setting.Email", "deeeet@gmail.com", "", nil},
	}, &UserProtoResponse{})

	if want := 2; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}
}

func TestUnmarshallerFunc(t *testing.T) {
	unmarshalFunc := protoUnmarshalFunc
	if err := unmarshalFunc([]byte(""), &User{}); err == nil {
		t.Fatal("expect to be failed")
	}
}

func TestAssertFunc(t *testing.T) {
	var buf bytes.Buffer
	tFatalf = fprintFatalFunc(&buf)
	defaultAssertFunc(t, 1, 2, "test-assert")
	if got, want := buf.String(), "test-assert: got 2(int), want 1(int)"; !strings.Contains(got, want) {
		t.Fatalf("expect %q to contain %q", got, want)
	}

}
