# go-httpdoc [![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godoc] [![Travis](https://img.shields.io/travis/mercari/go-httpdoc.svg?style=flat-square)][travis]

[godoc]: http://godoc.org/github.com/mercari/go-httpdoc
[travis]: https://travis-ci.org/mercari/go-httpdoc

`go-httpdoc` is a Golang package for generating API documentation from [`httptest`](https://golang.org/pkg/net/http/httptest/) test cases. 

It provides a simple http middleware for recording various http requst & response values you use in your tests and automatically arranges and generates them as usable documentation in markdown format (See the simple example output [here](/_example/doc/validate.md)). It also provides a way to validate values are equal to what you expect with annotation (e.g., you can add a description for headers, params or response fields). If you write proper tests, it generates usable documentation (namely, it forces you to write good tests). 

Not only json request & response but it supports [protocol buffer](https://developers.google.com/protocol-buffers/) (See example output of protocol buffer [here](/_example/doc/protobuf.md)).

See usage and example on [GoDoc](https://godoc.org/github.com/mercari/go-httpdoc).

*NOTE*: This package is experimental and may make backwards-incompatible changes.

## Install

Use go get,

```
$ go get github.com/mercari/go-httpdoc
```

## Usage

All usage is describe in [GoDoc](https://godoc.org/github.com/mercari/go-httpdoc). 

To generate documentation, you need to set the following env var,

```bash
$ export HTTPDOC=1
```

## Reference

The original ideas comes from [r7kamura/autodoc](https://github.com/r7kamura/autodoc) (rack middleware).

For struct inspection in validator, we use [tenntenn/gpath](https://github.com/tenntenn/gpath) package.
