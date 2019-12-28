package memfs

import (
	"encoding/binary"
	"errors"
	"log"
	"sync"
	"time"
	"unsafe"

	"nfs/xdrrpc/nfs"
)

type file struct {
	mu    sync.Mutex
	buf   []byte
	mtime int64
	ctime int64
}

func NewFile(content string) *file {
	return &file{
		buf:   []byte(content),
		mtime: time.Now().Unix(),
		ctime: time.Now().Unix(),
	}
}

func (f *file) ID() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(uintptr(unsafe.Pointer(f))))
	return b
}

func (f *file) Attr() nfs.Fattr3 {
	return nfs.Fattr3{
		Type:     nfs.NF3Reg,
		FileMode: 0644,
		Nlink:    1,
		UID:      0,
		GID:      0,
		Filesize: uint64(len(f.buf)),
		Used:     uint64(len(f.buf)),
		FSID:     83,
		Fileid:   uint64(uintptr(unsafe.Pointer(f))),
		Atime:    nfs.NFS3Time{Seconds: uint32(f.mtime)},
		Mtime:    nfs.NFS3Time{Seconds: uint32(f.mtime)},
		Ctime:    nfs.NFS3Time{Seconds: uint32(f.ctime)},
	}
}

func (t *file) Access(res *nfs.ACCESS3res) error {
	res.Status = nfs.NFSStatOk
	res.Access = 0x3f
	return nil
}

func (f *file) Getattr(res *nfs.GETATTR3res) error {
	res.Status = nfs.NFSStatOk
	res.Attr = f.Attr()
	return nil
}

func (f *file) Setattr(args *nfs.SETATTR3args, res *nfs.SETATTR3res) error {
	f.mtime = time.Now().Unix()
	res.Status = nfs.NFSStatOk
	if args.Sattr.Size.IsSet {
		length := args.Sattr.Size.Size
		if uint64(len(f.buf)) < length {
			new := make([]byte, length)
			copy(new, f.buf)
			f.buf = new
		}
		f.buf = f.buf[:length]
	}
	return nil
}

func (f *file) Read(args *nfs.READ3args, res *nfs.READ3res) error {
	if args.Offset >= uint64(len(f.buf)) {
		res.Status = nfs.NFSStatOk
		res.EOF = true
		return nil
	}
	res.Status = nfs.NFSStatOk
	res.Data = make([]byte, args.Count)
	n := copy(res.Data, f.buf[args.Offset:])
	if n < len(res.Data) {
		res.EOF = true
	}
	res.Count = uint32(n)

	return nil
}

func (f *file) Write(args *nfs.WRITE3args, res *nfs.WRITE3res) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.mtime = time.Now().Unix()

	count := uint32(len(args.Data))
	if count != args.Count {
		return errors.New("write size mismatch")
	}

	// append
	if args.Offset == uint64(len(f.buf)) {
		log.Println("append", args.Offset, uint64(len(f.buf)))
		f.buf = append(f.buf, args.Data...)
		res.Status = nfs.NFSStatOk
		res.Count = args.Count
		res.Committed = 2
		return nil
	}

	// optionally realloc
	length := args.Offset + uint64(count)
	if uint64(len(f.buf)) < length {
		if uint64(cap(f.buf)) < length {
			new := make([]byte, int64(float64(length)*1.05))
			copy(new, f.buf)
			f.buf = new
		}
		f.buf = f.buf[:length]
	}

	// memcopy
	copy(f.buf[args.Offset:], args.Data)
	res.Status = nfs.NFSStatOk
	res.Count = args.Count
	res.Committed = 2

	return nil
}

func (f *file) Commit(args *nfs.COMMIT3args, res *nfs.COMMIT3res) error {
	f.mtime = time.Now().Unix()
	res.Status = nfs.NFSStatOk
	return nil
}
