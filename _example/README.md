# httpdoc example

This directory contains some example of `httpdoc`. 

- [`handler_simple_test.go`](/_example/handler_simple_test.go) generates [`doc/simple.md`](/_example/doc/simple.md)
- [`handler_validate_test.go`](/_example/handler_validate_test.go) generates [`doc/validate.md`](/_example/doc/validate.md)
- [`handler_proto_test.go`](/_example/handler_proto_test.go) generates [`doc/protobuf.md`](/_example/doc/protobuf.md)

To generate document by yourself, run the following command,

```bash
$ env HTTPDOC=1 go test -v
```

One example uses protocol buffer, message defenition is in `../proto` directory. To generate code from that, run the following command,

```bash
$ protoc -I=./../proto --gofast_out=./ ../proto/message.proto
```
