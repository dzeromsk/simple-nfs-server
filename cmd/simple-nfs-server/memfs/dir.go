package memfs

import (
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/dzeromsk/xdrrpc/nfs"
)

type Node interface {
	ID() []byte
	Attr() nfs.Fattr3
}

type dir struct {
	nodes map[string]Node
	mux   nfs.ServeMux
}

func NewDir(mux nfs.ServeMux, nodes map[string]Node) *dir {
	d := &dir{
		mux:   mux,
		nodes: nodes,
	}
	// d.nodes[".."] = ?
	d.nodes["."] = d
	return d
}

func (d *dir) ID() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(uintptr(unsafe.Pointer(d))))
	return b
}

func (d *dir) Attr() nfs.Fattr3 {
	const size = uint64(unsafe.Sizeof(*d))
	return nfs.Fattr3{
		Type:     nfs.NF3Dir,
		FileMode: 0755 | uint32(os.ModeDir),
		Nlink:    1,
		UID:      0,
		GID:      0,
		Filesize: size,
		Used:     size,
		FSID:     83,
		Fileid:   uint64(uintptr(unsafe.Pointer(d))),
		Atime:    starttime,
		Mtime:    starttime,
		Ctime:    starttime,
	}
}

func (d *dir) Access(res *nfs.ACCESS3res) error {
	res.Status = nfs.NFSStatOk
	res.Access = 0x3f
	res.Attr.IsSet = true
	res.Attr.Attr = d.Attr()

	return nil
}

func (d *dir) Getattr(res *nfs.GETATTR3res) error {
	res.Status = nfs.NFSStatOk
	res.Attr = d.Attr()
	return nil
}

func (d *dir) Readdirplus(args *nfs.READDIRPLUS3args, res *nfs.READDIRPLUS3res) error {
	var prev *nfs.Entryplus3

	var limit = args.MaxCount / uint32(unsafe.Sizeof(*prev))
	if limit > args.DirCount {
		limit = args.DirCount
	}

	// TODO(dzeromsk): add support for cookies, create slice
	// of names, sort it, iterate over it and use index as
	// a cookie
	var n uint64 = 1
	for filename, node := range d.nodes {
		res.Reply.Entry = &nfs.Entryplus3{
			Cookie:   n,
			FileID:   n,
			FileName: filename,
			Next:     prev,
			Handle: nfs.PostOpFH3{
				IsSet: true,
				FH:    node.ID(),
			},
			Attr: nfs.PostOpAttr{
				IsSet: true,
				Attr:  node.Attr(),
			},
		}
		prev = res.Reply.Entry

		if n > uint64(limit) {
			break
		}
		n++
	}
	res.Reply.EOF = true

	res.Status = nfs.NFSStatOk
	// removes second readdir
	res.CookieVerf = uint64(n)
	res.Attr.IsSet = true
	res.Attr.Attr = d.Attr()

	return nil
}

func (d *dir) Mkdir(name string, attr *nfs.Sattr3, res *nfs.MKDIR3res) error {
	new := NewDir(d.mux, map[string]Node{
		"..": d,
	})

	id := new.ID()

	d.mux.Handle(id, new)
	d.nodes[name] = new

	// xxx.nodes[name] = h
	res.Handle.IsSet = true
	res.Handle.FH = id
	res.Attr.IsSet = true
	res.Attr.Attr = new.Attr()
	//res.DirWcc
	return nil
}

func (d *dir) Create(name string, res *nfs.CREATE3res) error {
	new := NewFile("")

	id := new.ID()

	d.mux.Handle(id, new)
	d.nodes[name] = new

	res.Handle.IsSet = true
	res.Handle.FH = id
	res.Attr.IsSet = true
	res.Attr.Attr = new.Attr()

	return nil
}

func (d *dir) Lookup(name string, res *nfs.LOOKUP3res) error {
	node, ok := d.nodes[name]
	if !ok {
		res.Status = nfs.NFSStatNoent
		return nil
	}

	id := node.ID()

	d.mux.Handle(id, node)

	res.Status = nfs.NFSStatOk
	res.Object = id
	res.DirAttr.IsSet = true
	res.DirAttr.Attr = d.Attr()
	res.Attr.IsSet = true
	res.Attr.Attr = node.Attr()
	return nil
}

func (d *dir) Link(object []byte, name string, res *nfs.LINK3res) error {
	o, ok := d.mux.Load(object)
	if !ok {
		res.Status = nfs.NFSStatStale
		return nil
	}

	node, ok := o.(Node)
	if !ok {
		res.Status = nfs.NFSStatStale
		return nil
	}

	d.nodes[name] = node

	res.Status = nfs.NFSStatOk
	res.Attr.IsSet = true
	res.Attr.Attr = node.Attr()

	return nil
}

func (d *dir) Remove(name string, res *nfs.REMOVE3res) error {
	node, ok := d.nodes[name]
	if !ok {
		res.Status = nfs.NFSStatNoent
		return nil
	}
	if _, ok := node.(*file); !ok {
		res.Status = nfs.NFSStatIsdir
		return nil
	}

	id := node.ID()

	d.mux.Delete(id)
	delete(d.nodes, name)

	return nil
}

func (d *dir) Rmdir(name string, res *nfs.RMDIR3res) error {
	node, ok := d.nodes[name]
	if !ok {
		res.Status = nfs.NFSStatNoent
		return nil
	}
	if _, ok := node.(*dir); !ok {
		res.Status = nfs.NFSStatNotdir
		return nil
	}

	id := node.ID()

	d.mux.Delete(id)
	delete(d.nodes, name)

	return nil
}

func (d *dir) Setattr(args *nfs.SETATTR3args, res *nfs.SETATTR3res) error {
	res.Status = nfs.NFSStatOk
	return nil
}

func (d *dir) Rename(args *nfs.RENAME3args, res *nfs.RENAME3res) error {
	from, ok := d.nodes[args.From.Name]
	if !ok {
		res.Status = nfs.NFSStatNoent
		return nil
	}
	node, ok := d.mux.Load(args.To.Dir)
	if !ok {
		res.Status = nfs.NFSStatInval
		return nil
	}
	dir, ok := node.(*dir)
	if !ok {
		// support for top level renames
		fs, ok2 := node.(*fs)
		if !ok2 {
			res.Status = nfs.NFSStatNotdir
			return nil
		}
		dir = fs.dir
	}

	// If the directory, to.dir, already contains an entry with
	// the name, to.name, the source object must be compatible
	// with the target: either both are non-directories or both
	// are directories and the target must be empty. If
	// compatible, the existing target is removed before the
	// rename occurs. If they are not compatible or if the target
	// is a directory but not empty, the server should return the
	// error, NFS3ERR_EXIST.
	//
	// _, ok = dir.nodes[args.To.Name]
	// if ok {
	// 	res.Status = nfs.NFSStatExist
	// 	return nil
	// }

	// add node to dst dir
	dir.nodes[args.To.Name] = from

	// delete file from src dir
	delete(d.nodes, args.From.Name)

	res.Status = nfs.NFSStatOk
	return nil
}
