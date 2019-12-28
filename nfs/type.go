package nfs

const (
	Nfs3Prog = 100003
	Nfs3Vers = 3

	// The size in bytes of the opaque cookie verifier passed by
	// READDIR and READDIRPLUS.
	NFS3_COOKIEVERFSIZE = 8

	// file types
	NF3Reg  = 1
	NF3Dir  = 2
	NF3Blk  = 3
	NF3Chr  = 4
	NF3Lnk  = 5
	NF3Sock = 6
	NF3FIFO = 7
)

type AuthFlavor int32

const (
	AuthNone  AuthFlavor = iota // No authentication
	AuthSys                     // Unix style (uid+gids)
	AuthShort                   // Short hand unix style
	AuthDh                      // DES style (encrypted timestamp)
	AuthKerb                    // Keberos Auth
	AuthRSA                     // RSA authentication
	RPCsecGss                   // GSS-based RPC security
)

// TODO(dzeromsk): proepr definition
type NullArgs struct{}

type NullRes struct{}

type NFSStat int32

const (
	NFSStatOk          NFSStat = 0
	NFSStatPerm        NFSStat = 1
	NFSStatNoent       NFSStat = 2
	NFSStatIo          NFSStat = 5
	NFSStatNxio        NFSStat = 6
	NFSStatAcces       NFSStat = 13
	NFSStatExist       NFSStat = 17
	NFSStatXdev        NFSStat = 18
	NFSStatNodev       NFSStat = 19
	NFSStatNotdir      NFSStat = 20
	NFSStatIsdir       NFSStat = 21
	NFSStatInval       NFSStat = 22
	NFSStatFbig        NFSStat = 27
	NFSStatNospc       NFSStat = 28
	NFSStatRofs        NFSStat = 30
	NFSStatMlink       NFSStat = 31
	NFSStatNametoolong NFSStat = 63
	NFSStatNotempty    NFSStat = 66
	NFSStatDquot       NFSStat = 69
	NFSStatStale       NFSStat = 70
	NFSStatRemote      NFSStat = 71
	NFSStatBadhandle   NFSStat = 10001
	NFSStatNotsync     NFSStat = 10002
	NFSStatBadcookie   NFSStat = 10003
	NFSStatNotsupp     NFSStat = 10004
	NFSStatToosmall    NFSStat = 10005
	NFSStatServerfault NFSStat = 10006
	NFSStatBadtype     NFSStat = 10007
	NFSStatJukebox     NFSStat = 10008
)

type PostOpAttr struct {
	IsSet bool   `xdr:"union"`
	Attr  Fattr3 `xdr:"unioncase=1"`
}

type NFS3Time struct {
	Seconds  uint32
	Nseconds uint32
}

type Fattr3 struct {
	Type     uint32
	FileMode uint32
	Nlink    uint32
	UID      uint32
	GID      uint32
	Filesize uint64
	Used     uint64
	SpecData [2]uint32
	FSID     uint64
	Fileid   uint64
	Atime    NFS3Time
	Mtime    NFS3Time
	Ctime    NFS3Time
}

// TODO(dzeromsk): replace `union`s with `optional` if possible
type Sattr3Mode struct {
	IsSet bool   `xdr:"union"`
	Mode  uint32 `xdr:"unioncase=1"`
}

type Sattr3UID struct {
	IsSet bool   `xdr:"union"`
	UID   uint32 `xdr:"unioncase=1"`
}

type Sattr3GID struct {
	IsSet bool   `xdr:"union"`
	GID   uint32 `xdr:"unioncase=1"`
}

type Sattr3Size struct {
	IsSet bool   `xdr:"union"`
	Size  uint64 `xdr:"unioncase=1"`
}

type Sattr3Time struct {
	TimeHow int32    `xdr:"union"`
	Time    NFS3Time `xdr:"unioncase=2"`
}

type Sattr3 struct {
	Mode  Sattr3Mode
	UID   Sattr3UID
	GID   Sattr3GID
	Size  Sattr3Size
	Atime Sattr3Time
	Mtime Sattr3Time
}

type FSINFO3args struct {
	Object []byte
}

type FSINFO3res struct {
	Status     NFSStat
	Attr       PostOpAttr
	RTMax      uint32
	RTPref     uint32
	RTMult     uint32
	WTMax      uint32
	WTPref     uint32
	WTMult     uint32
	DTPref     uint32
	Size       uint64
	TimeDelta  NFS3Time
	Properties uint32
}

type PATHCONF3args struct {
	Object []byte
}

type PATHCONF3res struct {
	Status          NFSStat
	Attr            PostOpAttr
	LinkMax         uint32
	NameMax         uint32
	NoTrunc         bool
	ChownRestricted bool
	CaseInsensitive bool
	CaseIreserving  bool
}

type GETATTR3args struct {
	Object []byte
}

type GETATTR3res struct {
	Status NFSStat
	Attr   Fattr3
}

type ACCESS3args struct {
	Object []byte
	Access uint32
}

type ACCESS3res struct {
	Status NFSStat
	Attr   PostOpAttr
	Access uint32
}

type FSSTAT3args struct {
	FSRoot []byte
}

type FSSTAT3res struct {
	Status   NFSStat
	Attr     PostOpAttr
	Tbytes   uint64
	Fbytes   uint64
	Abytes   uint64
	Tfiles   uint64
	Ffiles   uint64
	Afiles   uint64
	Invarsec uint32
}

type Diropargs3 struct {
	Dir  []byte
	Name string
}

type LOOKUP3args struct {
	What Diropargs3
}

type LOOKUP3res struct {
	Status  NFSStat
	Object  []byte
	Attr    PostOpAttr
	DirAttr PostOpAttr
}

type READDIRPLUS3args struct {
	Dir        []byte
	Cookie     uint64
	CookieVerf uint64
	DirCount   uint32
	MaxCount   uint32
}

type READDIRPLUS3res struct {
	Status     NFSStat
	Attr       PostOpAttr
	CookieVerf uint64
	Reply      DirListPlus3
}

type DirListPlus3 struct {
	Entry *Entryplus3 `xdr:"optional"`
	EOF   bool
}

type Entryplus3 struct {
	FileID   uint64
	FileName string
	Cookie   uint64
	Attr     PostOpAttr
	Handle   PostOpFH3
	Next     *Entryplus3 `xdr:"optional"`
}

type PostOpFH3 struct {
	IsSet bool   `xdr:"union"`
	FH    []byte `xdr:"unioncase=1"`
}

type READ3args struct {
	Object []byte
	Offset uint64
	Count  uint32
}

type READ3res struct {
	Status NFSStat
	Attr   PostOpAttr
	Count  uint32
	EOF    bool
	Data   []byte
}

type WccAttr struct {
	Size  uint64
	Mtime NFS3Time
	Ctime NFS3Time
}

type PreOpAttr struct {
	IsSet bool    `xdr:"union"`
	Attr  WccAttr `xdr:"unioncase=1"`
}

type WccData struct {
	Pre  PreOpAttr
	Post PostOpAttr
}

type MKDIR3args struct {
	Where Diropargs3
	Attr  Sattr3
}

type MKDIR3res struct {
	Status NFSStat
	Handle PostOpFH3
	Attr   PostOpAttr
	DirWcc WccData
}

type Createhow3 struct {
	Mode          int32   `xdr:"union"`
	UncheckedAttr Sattr3  `xdr:"unioncase=0"`
	GuardedAttr   Sattr3  `xdr:"unioncase=1"`
	CreateVerf    [8]byte `xdr:"unioncase=2"`
}

type CREATE3args struct {
	Where Diropargs3
	How   Createhow3
}

type CREATE3res struct {
	Status NFSStat
	Handle PostOpFH3
	Attr   PostOpAttr
	DirWcc WccData
}

type Sattrguard3 struct {
	IsSet bool     `xdr:"union"`
	Ctime NFS3Time `xdr:"unioncase=1"`
}

type SETATTR3args struct {
	Object []byte
	Sattr  Sattr3
	Guard  Sattrguard3
}

type SETATTR3res struct {
	Status NFSStat
	ObjWcc WccData
}

type LINK3args struct {
	Object []byte
	Link   Diropargs3
}

type LINK3res struct {
	Status NFSStat
	Attr   PostOpAttr
	DirWcc WccData
}

type REMOVE3args struct {
	Object Diropargs3
}

type REMOVE3res struct {
	Status NFSStat
	DirWcc WccData
}

type RMDIR3args struct {
	Object Diropargs3
}

type RMDIR3res struct {
	Status NFSStat
	DirWcc WccData
}

type WRITE3args struct {
	Object []byte
	Offset uint64
	Count  uint32
	Stable int32
	Data   []byte
}

type WRITE3res struct {
	Status    NFSStat
	FileWcc   WccData
	Count     uint32
	Committed int32
	Verf      [8]byte
}

type COMMIT3args struct {
	Object []byte
	Offset uint64
	Count  uint32
}

type COMMIT3res struct {
	Status  NFSStat
	FileWcc WccData
	Verf    [8]byte
}

type RENAME3args struct {
	From Diropargs3
	To   Diropargs3
}

type RENAME3res struct {
	Status     NFSStat
	FromDirWcc WccData
	ToDirWcc   WccData
}
