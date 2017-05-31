// Package httpdoc is a Golang package for generating API documentation from httptest test cases.
//
// It provides a simple http middleware for recording various http requst & response values you use in your tests
// and automatically arranges and generates them as usable documentation in markdown format. It also provides a way
// to validate values are equal to what you expect with annotation (e.g., you can add a description for headers,
// params or response fields).
//
// See example document output, https://github.com/mercari/go-httpdoc/blob/master/_example/doc/validate.md
package httpdoc
