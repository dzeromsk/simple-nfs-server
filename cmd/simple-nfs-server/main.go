package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"runtime"

	"github.com/dzeromsk/xdrrpc"
	"github.com/dzeromsk/xdrrpc/nfs"

	"github.com/dzeromsk/xdrrpc/cmd/simple-nfs-server/memfs"
)

var (
	listen = flag.String("listen", ":12049", "Server listen address")
	debug  = flag.Bool("debug", false, "Enable debug prints")
)

func main() {
	flag.Parse()

	xdrrpc.Debug = *debug

	root := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}

	mount := memfs.NewMount(root)
	rpc.Register(mount)

	var mux = nfs.NewServeMux()
	mux.Handle(root, memfs.NewFS(
		memfs.NewDir(mux, map[string]memfs.Node{
			"hello": memfs.NewFile("world\n"),
			"foo":   memfs.NewFile("bar\n"),
			"example": memfs.NewDir(mux, map[string]memfs.Node{
				"alice": memfs.NewFile("bob\n"),
			}),
		}),
	))
	rpc.Register(mux.Receiver())

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalln("listen error:", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln("accept error:", err)
		}
		go func(conn net.Conn) {
			defer func() {
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					log.Printf("example: panic serving %s: %v\n%s", conn.RemoteAddr(), err, buf)
				}
				conn.Close()
			}()

			xdrrpc.ServeConn(conn)
		}(conn)
	}
}
