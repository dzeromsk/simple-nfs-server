package memfs

import (
	"github.com/dzeromsk/xdrrpc/mount"
	"github.com/dzeromsk/xdrrpc/nfs"
)

type Mount struct {
	root []byte
}

func NewMount(root []byte) *Mount {
	return &Mount{root: root}
}

func (m *Mount) Null(args *mount.NullArgs, res *mount.NullRes) error {
	return nil
}

func (m *Mount) Mount(args *mount.MountArgs, res *mount.MountRes) error {
	res.Handle = m.root
	res.AuthFlavors = []nfs.AuthFlavor{nfs.AuthSys}
	return nil
}
