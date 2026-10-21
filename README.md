# ifc
Interface of docker container like `ifconfig` or `ip`

## Build

Rebuild the executable with Go 1.26.10 (or a newer patched supported
toolchain) for the `crypto/x509` fix in
[GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809). Existing binaries must
be rebuilt; updating this repository does not update installed executables.

```sh
GOTOOLCHAIN=go1.26.10 go test -mod=readonly -race ./...
GOTOOLCHAIN=go1.26.10 go build -mod=readonly -trimpath -o ifc .
go version -m ./ifc
```

The `go 1.14` directive retains the existing language version. The toolchain
directive selects the patched compiler for local module builds, and CI pins
and verifies that compiler independently through `.go-version`.
