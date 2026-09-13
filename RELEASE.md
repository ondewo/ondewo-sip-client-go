# Release History

*****************

## Release ONDEWO SIP Go Client 0.1.0

### New Features

* Initial release of the ONDEWO SIP (SIP Telephony) gRPC client for Go. The module
  ships the stubs generated from the [ONDEWO SIP API](https://github.com/ondewo/ondewo-sip-api)
  by version 5.15.0 of the
  [ONDEWO Proto Compiler](https://github.com/ondewo/ondewo-proto-compiler): one `*.pb.go` of
  messages and one `*_grpc.pb.go` of service stubs per `.proto`, below `api/ondewo/sip/`,
  compiled against the `google.golang.org/protobuf` and `google.golang.org/grpc` runtimes pinned by
  the compiler image.
* `make build` reproduces the whole client from the two submodules — proto compiler image, stub
  generation and `go build` — and `make check_build` asserts that every `.proto` of the API
  produced a stub.

### Improvements

* The release targets tag each release twice on the same commit: with the ONDEWO release number
  (`0.1.0`) that the rest of the fleet uses, and with the `v`-prefixed spelling (`v0.1.0`) that is
  the only tag shape the Go module resolver accepts. `make publish_go_module` then warms
  `proxy.golang.org` so the new version is immediately installable with `go get`.
