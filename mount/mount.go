package mount

import (
	"github.com/dzeromsk/xdrrpc"
	"github.com/dzeromsk/xdrrpc/nfs"
)

func init() {
	xdrrpc.Register(100005, 3, 0, "Mount", "Null")
	xdrrpc.Register(100005, 3, 1, "Mount", "Mount")
	xdrrpc.Register(100005, 3, 2, "Mount", "Dump")
	xdrrpc.Register(100005, 3, 3, "Mount", "Unmount")
	xdrrpc.Register(100005, 3, 4, "Mount", "UnmountAll")
	xdrrpc.Register(100005, 3, 5, "Mount", "Export")
}

type NullArgs struct{}

type NullRes struct{}

type MountArgs struct {
	Dirpath string
}

type MountRes struct {
	Status      int32
	Handle      []byte
	AuthFlavors []nfs.AuthFlavor
}
