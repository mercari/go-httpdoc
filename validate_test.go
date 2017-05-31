package httpdoc

import (
	"reflect"
	"testing"

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
	}

	validator.RequestParams(t, []TestCase{
		{"token", "12345", ""},
		{"pretty", "true", ""},
	})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestParams(t, []TestCase{
		{"token", "8976", ""},
		{"pretty", "", ""},
		{"id", "u8988", ""},
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
		{"User-Agent", "Googlebot/2.1", ""},
		{"Content-Type", "application/json", ""},
		{"X-API-Version", "1.1.2", ""},
	})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestHeaders(t, []TestCase{
		{"User-Agent", []string{"curl"}, ""},
		{"Content-Type", []string{"application/protobuf"}, ""},
		{"X-API-Version", []string{"3.0"}, ""},
	})
	if want := 3; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
	}
}

func TestValidator_ResponseHeaders(t *testing.T) {
	validator := newValidator()
	validator.record.responseHeaders = map[string][]string{
		"Content-Type":  []string{"application/json"},
		"X-API-Version": []string{"1.1.2"},
	}
	validator.ResponseHeaders(t, []TestCase{
		{"Content-Type", "application/json", ""},
		{"X-API-Version", "1.1.2", ""},
	})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseHeaders(t, []TestCase{
		{"Content-Type", []string{"application/protobuf"}, ""},
	})
	if want := 1; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
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
		{"ID", 910, ""},
		{"Setting.Email", "taichi@mercari.com", ""},
	}, &User{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestBody(t, []TestCase{
		{"ID", 123, ""},
		{"Active", true, ""},
		{"Setting.Email", "deeeet@gmail.com", ""},
	}, &User{})

	if want := 3; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
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
	validator.ResponseBody(t, []TestCase{
		{"ID", 789, ""},
		{"Active", false, ""},
		{"Setting.Email", "tcnksm@mercari.com", ""},
		{"Permission[1]", "read", ""},
		{`Preference["email"]`, 0, ""},
	}, &User{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseBody(t, []TestCase{
		{"ID", 123, ""},
		{"Active", true, ""},
		{"Setting.Email", "deeeet@gmail.com", ""},
		{"Permission[1]", "write", ""},
		{`Preference["email"]`, 1, ""},
	}, &User{})

	if want := 5; got != want {
		t.Fatalf("expect valiate fails %d, got %d", want, got)
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
	validator := newValidator()
	validator.validateFields(t, []TestCase{
		{"ID", 12345, ""},
		{"Name", "tcnksm", ""},
		{"Active", true, ""},
		{"Setting.Email", "tcnksm@example.com", ""},
		{"Setting.SNS.Twitter", "@deeeet", ""},
		{"Permission[0]", "write", ""},
		{`Preference["email"]`, 0, ""},
	}, testUser, &[]Data{})
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
	validator.RequestBody(t, []TestCase{
		{"Id", int32(12345), ""},
		{"Name", "tcnksm", ""},
	}, &UserProtoRequest{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.RequestBody(t, []TestCase{
		{"Id", 123, ""},
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
		{"Id", int32(667854), ""},
		{"Setting.Email", "httpdoc@example.com", ""},
	}, &UserProtoResponse{})

	var got int
	validator.assertFunc = testAssertWithCount(&got)
	validator.ResponseBody(t, []TestCase{
		{"Id", 123, ""},
		{"Setting.Email", "deeeet@gmail.com", ""},
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
