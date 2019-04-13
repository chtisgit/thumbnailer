package thmb

import (
	"bytes"
	"io"
	"math"
	"net"
	"os"
)

// Thmb is a client that connects to a Thumbnailer Server. It can resize images for you.
type Thmb struct {
	nw, addr string
	conn     net.Conn
	MaxSize  uint32
}

const defaultMaxSize = 1024 * 1024 // 1 MB

// NewThmb creates a Thmb. The parameters are as in net.Dial.
func NewThmb(network, addr string) *Thmb {
	return &Thmb{
		nw:      network,
		addr:    addr,
		MaxSize: defaultMaxSize,
	}
}

func (t *Thmb) reinit() (err error) {
	if t.conn == nil {
		t.conn, err = net.Dial(t.nw, t.addr)
		return
	}
	return
}

// Close releases the resources of the Thmb.
func (t *Thmb) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}

// ResizeReader resizes an image read from an io.Reader. The size in bytes must be supplied.
func (t *Thmb) ResizeReader(r io.Reader, w io.Writer, size, width, height uint32) error {
	if size > math.MaxUint32 {
		panic("image file size too large")
	}

	err := t.reinit()
	if err != nil {
		return err
	}

	req := &Request{
		Size:   size,
		Width:  width,
		Height: height,
	}

	if err := req.Send(t.conn, r); err != nil {
		t.conn.Close()
		t.conn = nil
		return err
	}

	_, err = ReceiveResponse(t.conn, w, t.MaxSize)
	if err != nil {
		t.conn.Close()
		t.conn = nil
	}

	return err

}

// Resize resizes an image read from byte slice.
func (t *Thmb) Resize(img []byte, width, height uint32) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := t.ResizeReader(bytes.NewReader(img), buf, uint32(len(img)), width, height)
	return buf.Next(int(t.MaxSize)), err
}

// ResizeFile resizes an image read from a file.
func (t *Thmb) ResizeFile(path string, w io.Writer, width, height uint32) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.ResizeReader(f, w, uint32(st.Size()), width, height)
}
