package memfs

import (
	"time"

	"nfs/xdrrpc/nfs"
)

var starttime = nfs.NFS3Time{Seconds: uint32(time.Now().Unix())}

type fs struct {
	*dir
}

func NewFS(dir *dir) *fs {
	return &fs{dir: dir}
}

func (f *fs) Fsinfo(res *nfs.FSINFO3res) error {
	res.Status = nfs.NFSStatOk
	res.RTMax = 1048576
	res.RTPref = 1048576
	res.RTMult = 4096
	res.WTMax = 1048576
	res.WTPref = 1048576
	res.WTMult = 4096
	// res.DTPref = 4096
	res.DTPref = 32768 // max on linux
	res.Size = 17592186040320
	res.TimeDelta = nfs.NFS3Time{Seconds: 1}
	res.Properties = 0x0000001b
	return nil
}

func (f *fs) Fsstat(res *nfs.FSSTAT3res) error {
	res.Status = nfs.NFSStatOk
	res.Tbytes = 1 * 1024 * 1024 * 1024 * 1024 * 1024
	res.Fbytes = 0.5 * 1024 * 1024 * 1024 * 1024 * 1024
	res.Abytes = 0.5 * 1024 * 1024 * 1024 * 1024 * 1024
	res.Tfiles = 1024
	res.Ffiles = 512
	res.Afiles = 512
	res.Invarsec = 0
	return nil
}

func (f *fs) Pathconf(res *nfs.PATHCONF3res) error {
	res.Status = nfs.NFSStatOk
	return nil
}
