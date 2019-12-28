package xdrrpc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/rpc"
	"sync"

	"github.com/rasky/go-xdr/xdr2"
)

var (
	errInvalidMessageType = errors.New("xdrrpc: invalid message type received")
	errEncodingResponse   = errors.New("xdrrpc: xdr error encoding response")
	errMappingDuplicate   = errors.New("xdrrpc: service already defined")
)

var Debug = false

type serverCodec struct {
	dec *xdr.Decoder // for reading XDR values
	enc *xdr.Encoder // for writing XDR values
	c   io.WriteCloser

	lr  *io.LimitedReader // to limit decoder
	buf *bytes.Buffer     // for encoder

	// temporary work space
	req serverRequest
}

// NewServerCodec returns a new rpc.ServerCodec using JSON-RPC on conn.
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	// TODO(dzeromsk): we should wrap conn reader with bufio
	lr := &io.LimitedReader{R: conn, N: 4}
	buf := new(bytes.Buffer)
	return &serverCodec{
		lr:  lr,
		dec: xdr.NewDecoder(lr),
		enc: xdr.NewEncoder(buf),
		c:   conn,
		buf: buf,
	}
}

type serverRequest struct {
	Xid        uint32
	Type       MessageType
	RPCVersion uint32     // usually equal to 2
	Program    uint32     // Remote program
	Version    uint32     // Remote program's version
	Procedure  uint32     // Procedure number
	Cred       OpaqueAuth // Authentication credential
	Verf       OpaqueAuth // Authentication verifier
}

var emptyRequest = serverRequest{}

func (r *serverRequest) reset() {
	*r = emptyRequest
}

// We are assuming all responses are of type Accept
type serverResponse struct {
	Xid        uint32
	Type       MessageType
	ReplayStat ReplyStat
	Verf       OpaqueAuth
	Stat       AcceptStat
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	// reset limited reader
	c.lr.N = 4

	size, _, err := c.dec.DecodeUint()
	if err != nil {
		return err
	}

	// TODO(dzeromsk): we should be able to read multiple chained
	// records and not just assume there is always one

	// reset limited reader, now to proper size
	c.lr.N = int64(size & 0x7fffffff)

	c.req.reset()
	if _, err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	if c.req.Type != Call {
		return errInvalidMessageType
	}

	// We return "unknown" service and/or program because we want
	// rpc server to consume request body and return error message
	// to the client, simple return err will just exit codec
	name, ok := lookup(c.req.Program, c.req.Version, c.req.Procedure)
	if !ok {
		name = "unknown.unknown"
	}

	r.ServiceMethod = name
	r.Seq = uint64(c.req.Xid)

	if Debug {
		log.Printf("method: %s\n", name)
	}

	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	// rpc server will try to discard body by calling us with nil x
	if x == nil {
		_, err := io.Copy(ioutil.Discard, c.lr)
		return err
	}

	_, err := c.dec.Decode(x)

	// log.Printf("request: %s", spew.Sdump(x))

	// make sure we ignore ramaining data
	if _, err2 := io.Copy(ioutil.Discard, c.lr); err2 != nil {
		if err != nil && err2 != io.EOF {
			return err2
		}
	}

	return err
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	resp := serverResponse{
		Xid:        uint32(r.Seq),
		Type:       Reply,
		ReplayStat: MessageAccepted,
	}

	if r.Error == "" {
		resp.Stat = Success
	} else {
		// TODO(dzeromsk): user c.req to determine proper error code or
		// replace c.req with err error property
		resp.Stat = ProcUnavail
	}

	c.buf.Reset()

	// encode dummy size (placeholder)
	// TODO(dzeromsk): seek?
	prefix, err := c.enc.EncodeInt(0)
	if err != nil {
		return errEncodingResponse
	}

	// encode header
	if _, err := c.enc.Encode(resp); err != nil {
		return errEncodingResponse
	}

	// encode result
	if _, err := c.enc.Encode(x); err != nil {
		return errEncodingResponse
	}

	data := c.buf.Bytes()

	// fix size
	size := uint32(len(data)-prefix) | 0x80000000
	binary.BigEndian.PutUint32(data[:4], size)

	_, err = c.c.Write(data)
	return err
}

func (c *serverCodec) Close() error {
	return c.c.Close()
}

// ServeConn runs the XRP-RPC server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}

type key struct {
	Program   uint32
	Version   uint32
	Procedure uint32
}

var DefaultMap sync.Map

func Register(program, version, procedure uint32, service, method string) {
	key := key{
		Program:   program,
		Version:   version,
		Procedure: procedure,
	}
	if _, dup := DefaultMap.LoadOrStore(key, service+"."+method); dup {
		panic(errMappingDuplicate)
	}
}

func lookup(program, version, procedure uint32) (string, bool) {
	key := key{
		Program:   program,
		Version:   version,
		Procedure: procedure,
	}
	v, ok := DefaultMap.Load(key)
	if !ok {
		return "", false
	}
	return v.(string), true
}

type MessageType int32

const (
	Call  MessageType = 0
	Reply MessageType = 1
)

type AuthFlavor int32

// OpaqueAuth is a structure with AuthFlavor enumeration followed by up to
// 400 bytes that are opaque to (uninterpreted by) the RPC protocol
// implementation.
type OpaqueAuth struct {
	Flavor AuthFlavor
	Body   []byte
}

type AcceptStat int32

const (
	Success      AcceptStat = iota // RPC executed successfully
	ProgUnavail                    // Remote hasn't exported the program
	ProgMismatch                   // Remote can't support version number
	ProcUnavail                    // Program can't support procedure
	GarbageArgs                    // Procedure can't decode params
	SystemError                    // Other errors
)

type ReplyStat int32

const (
	MessageAccepted ReplyStat = 0
	MessageDenied   ReplyStat = 1
)
