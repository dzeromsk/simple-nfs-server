# xdrrpc (simple-nfs-server) - nfs server the easy way

## Overview

`simple-nfs-server` is an example nfs server implementation that serves files from memory. `xdrrpc` is a golang library to create XDRRPC services also known as SUNRPC or ONCRPC ([RFC 5531](https://tools.ietf.org/html/rfc5531)). Both are written in pure go, without using rpcgen toolchain (Sun Microsystems ONC RPC generator). `xdrrpc` conforms to golang `net/rpc` [ServerCodec](https://golang.org/pkg/net/rpc/#ServerCodec) interface. API documentation for `xdrrpc` can be found in [![GoDoc](https://godoc.org/github.com/dzeromsk/xdrrpc?status.svg)](https://godoc.org/github.com/dzeromsk/xdrrpc). 

## Installation

```bash
$ go get -u github.com/dzeromsk/xdrrpc/...
```

This will make the `simple-nfs-server` tool available in `${GOPATH}/bin`, which by default means `~/go/bin`.

## Usage

`simple-nfs-server` serves all files from memory with limited support for attributes.

```
Usage of simple-nfs-server:
  -debug
        Enable debug prints
  -listen string
        Server listen address (default ":12049")
```

## Example

Start `simple-nfs-server`
```bash
$ simple-nfs-server --debug -listen :12049
```

Mount on the client using fixed protocol (tcp), port (12049) and any mount point on localhost:
```bash
$ sudo mount -vvv -o nfsvers=3,proto=tcp,port=12049,mountvers=3,mountport=12049,mountproto=tcp 127.0.0.1:/ /mnt/example
```

Low level usage of `xdrrpc` package
```go
import (
  "net/rpc"
  "github.com/dzeromsk/xdrrpc"
)

func init() {
  xdrrpc.Register(100005, 3, 0, "Mount", "Null")
}

type Mount struct {}

func (m *Mount) Null(args *mount.NullArgs, res *mount.NullRes) error {
	return nil
}

func main() {
	rpc.Register(mount)
	ln, _ := net.Listen("tcp", *listen)
	for {
		conn, _ := ln.Accept()
		go xdrrpc.ServeConn(conn)
	}
}
```

For helpers like `nfs.ServeMux` usage please take a look at `xdrrpc/nfs` and `xdrrpc/example/memfs` packages. Skimming through [RFC 1813](https://tools.ietf.org/html/rfc1813) will help too.

## Features

 - Simple.
 - Memory only.
 - Compatible with Linux kernel NFS Client.
 - Implements stdlib [ServerCodec](https://golang.org/pkg/net/rpc/#ServerCodec).

## Downsides

 - Some features are missing.

## Philosophy

`xdrrpc` conforms to golang `net/rpc` [ServerCodec](https://golang.org/pkg/net/rpc/#ServerCodec) interface. At some point in future we should consider replacing `net/rpc` with custom RPC server implementation.

I made this to debug Linux NFS Client attribute caching behavior at work. Feel free to fork it.


