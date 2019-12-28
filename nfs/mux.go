package nfs

import (
	"encoding/binary"
	"nfs/xdrrpc"
	"sync"
)

func init() {
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 0, "NFS", "Null")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 1, "NFS", "Getattr")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 2, "NFS", "Setattr")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 3, "NFS", "Lookup")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 4, "NFS", "Access")
	// xdrrpc.Register(Nfs3Prog, Nfs3Vers, 5, "NFS", "Readlink")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 6, "NFS", "Read")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 7, "NFS", "Write")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 8, "NFS", "Create")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 9, "NFS", "Mkdir")
	// xdrrpc.Register(Nfs3Prog, Nfs3Vers, 10, "NFS", "Symlink")
	// xdrrpc.Register(Nfs3Prog, Nfs3Vers, 11, "NFS", "Mknod")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 12, "NFS", "Remove")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 13, "NFS", "Rmdir")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 14, "NFS", "Rename")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 15, "NFS", "Link")
	// xdrrpc.Register(Nfs3Prog, Nfs3Vers, 16, "NFS", "Readdir")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 17, "NFS", "Readdirplus")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 18, "NFS", "Fsstat")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 19, "NFS", "Fsinfo")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 20, "NFS", "Pathconf")
	xdrrpc.Register(Nfs3Prog, Nfs3Vers, 21, "NFS", "Commit")
}

type Handle string

type ServeMux interface {
	Handle(object []byte, handler interface{})
	Load(object []byte) (handler interface{}, ok bool)
	Delete(object []byte)
	Receiver() interface{}
}

func NewServeMux() ServeMux {
	m := &serveMux{
	// objects: map[uint64]interface{}{},
	}
	m.rpc = NFS{m}
	return m
}

type serveMux struct {
	rpc     NFS
	objects sync.Map
}

func (mux *serveMux) Handle(object []byte, handler interface{}) {
	o := binary.LittleEndian.Uint64(object)
	mux.objects.Store(o, handler)
}

func (mux *serveMux) Load(object []byte) (handler interface{}, ok bool) {
	o := binary.LittleEndian.Uint64(object)
	return mux.objects.Load(o)
}

func (mux *serveMux) Delete(object []byte) {
	o := binary.LittleEndian.Uint64(object)
	mux.objects.Delete(o)
}

func (mux *serveMux) Receiver() interface{} {
	return &mux.rpc
}

type NFS struct {
	mux *serveMux
}

func (r *NFS) Null(args *NullArgs, res *NullRes) error {
	return nil
}

func (r *NFS) Fsinfo(args *FSINFO3args, res *FSINFO3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Fsinfo(*FSINFO3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Fsinfo(res)
}

func (r *NFS) Getattr(args *GETATTR3args, res *GETATTR3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Getattr(*GETATTR3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Getattr(res)
}

func (r *NFS) Access(args *ACCESS3args, res *ACCESS3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Access(*ACCESS3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Access(res)
}

func (r *NFS) Fsstat(args *FSSTAT3args, res *FSSTAT3res) error {
	node, ok := r.mux.Load(args.FSRoot)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Fsstat(*FSSTAT3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Fsstat(res)
}

func (r *NFS) Pathconf(args *PATHCONF3args, res *PATHCONF3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Pathconf(*PATHCONF3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Pathconf(res)
}

func (r *NFS) Lookup(args *LOOKUP3args, res *LOOKUP3res) error {
	node, ok := r.mux.Load(args.What.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Lookup(string, *LOOKUP3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Lookup(args.What.Name, res)
}

func (r *NFS) Readdirplus(args *READDIRPLUS3args, res *READDIRPLUS3res) error {
	node, ok := r.mux.Load(args.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Readdirplus(*READDIRPLUS3args, *READDIRPLUS3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Readdirplus(args, res)
}

func (r *NFS) Read(args *READ3args, res *READ3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Read(*READ3args, *READ3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Read(args, res)
}

func (r *NFS) Mkdir(args *MKDIR3args, res *MKDIR3res) error {
	node, ok := r.mux.Load(args.Where.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Mkdir(string, *Sattr3, *MKDIR3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Mkdir(args.Where.Name, &args.Attr, res)
}

func (r *NFS) Create(args *CREATE3args, res *CREATE3res) error {
	node, ok := r.mux.Load(args.Where.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Create(string, *CREATE3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Create(args.Where.Name, res)
}

func (r *NFS) Setattr(args *SETATTR3args, res *SETATTR3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Setattr(*SETATTR3args, *SETATTR3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Setattr(args, res)
}

func (r *NFS) Link(args *LINK3args, res *LINK3res) error {
	node, ok := r.mux.Load(args.Link.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Link([]byte, string, *LINK3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Link(args.Object, args.Link.Name, res)
}

func (r *NFS) Remove(args *REMOVE3args, res *REMOVE3res) error {
	node, ok := r.mux.Load(args.Object.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Remove(string, *REMOVE3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Remove(args.Object.Name, res)
}

func (r *NFS) Rmdir(args *RMDIR3args, res *RMDIR3res) error {
	node, ok := r.mux.Load(args.Object.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Rmdir(string, *RMDIR3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Rmdir(args.Object.Name, res)
}

func (r *NFS) Write(args *WRITE3args, res *WRITE3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Write(*WRITE3args, *WRITE3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Write(args, res)
}

func (r *NFS) Commit(args *COMMIT3args, res *COMMIT3res) error {
	node, ok := r.mux.Load(args.Object)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Commit(*COMMIT3args, *COMMIT3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Commit(args, res)
}

func (r *NFS) Rename(args *RENAME3args, res *RENAME3res) error {
	node, ok := r.mux.Load(args.From.Dir)
	if !ok {
		res.Status = NFSStatStale
		return nil
	}
	n, ok := node.(interface {
		Rename(*RENAME3args, *RENAME3res) error
	})
	if !ok {
		res.Status = NFSStatInval
		return nil
	}
	return n.Rename(args, res)
}
