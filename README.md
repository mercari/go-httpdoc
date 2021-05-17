# go-httpdoc [![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godoc] [![Travis](https://img.shields.io/travis/mercari/go-httpdoc/master.svg?style=flat-square)][travis]

[godoc]: http://godoc.org/go.mercari.io/go-httpdoc
[travis]: https://travis-ci.org/mercari/go-httpdoc

`go-httpdoc` is a Golang package to generate API documentation from [`httptest`](https://golang.org/pkg/net/http/httptest/) test cases.

It provides a simple http middleware which records http requests and responses from tests and generates documentation automatically in markdown format. See [Sample Documentation](/_example/doc/validate.md). It also provides a way to validate values are equal to what you expect with annotation (e.g., you can add a description for headers, params or response fields). If you write proper tests, it will generate usable documentation (namely, it forces you to write good tests).

Not only JSON request and response but it also supports [protocol buffer](https://developers.google.com/protocol-buffers/). See [Sample ProtoBuf Documentation](/_example/doc/protobuf.md)).

See usage and example in [GoDoc](https://godoc.org/go.mercari.io/go-httpdoc).

*NOTE*: This package is experimental and may make backward-incompatible changes.

## Prerequisites

go-httpdoc requires Go 1.7 or later.

## Install

Use go get:

```
$ go get -u go.mercari.io/go-httpdoc
```

## Usage

All usage are described in [GoDoc](https://godoc.org/go.mercari.io/go-httpdoc).

To generate documentation, set the following env var:

```bash
$ export HTTPDOC=1
```

## Reference

The original idea came from [r7kamura/autodoc](https://github.com/r7kamura/autodoc) (rack middleware).

For struct inspection in validator, it uses [tenntenn/gpath](https://github.com/tenntenn/gpath) package.
