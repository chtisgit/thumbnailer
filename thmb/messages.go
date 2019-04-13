package thmb

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"time"
)

// Request is sent by the client to the server, to request a thumbnail.
type Request struct {
	ID     uint64
	Size   uint32
	Width  uint32
	Height uint32

	// Server internals
	path string
	from net.Conn
}

// Response is sent from the server to the client to deliver the thumbnail image.
type Response struct {
	ID   uint64
	Data []byte

	// Server internals
	from net.Conn
}

// Send sends the request to the io.Writer. The image data is read from the io.Reader.
func (r *Request) Send(w io.Writer, src io.Reader) error {
	var buf [20]byte
	binary.BigEndian.PutUint64(buf[:], r.ID)
	binary.BigEndian.PutUint32(buf[8:], r.Size)
	binary.BigEndian.PutUint32(buf[12:], r.Width)
	binary.BigEndian.PutUint32(buf[16:], r.Height)

	if _, err := w.Write(buf[:]); err != nil {
		return err
	}

	_, err := io.CopyN(w, src, int64(r.Size))
	return err
}

// ReceiveRequest returns a Client's request which is read from the io.Reader.
func ReceiveRequest(r io.Reader, c *Config) (*Request, error) {
	var buf [20]byte

	_, err := r.Read(buf[:])
	if err != nil {
		return nil, err
	}

	req := &Request{
		ID:     binary.BigEndian.Uint64(buf[:]),
		Size:   binary.BigEndian.Uint32(buf[8:]),
		Width:  binary.BigEndian.Uint32(buf[12:]),
		Height: binary.BigEndian.Uint32(buf[16:]),
	}
	if req.Size > c.MaxSize {
		return nil, fmt.Errorf("request payload too large (%d bytes)", req.Size)
	}

	req.path = fmt.Sprintf("%s/%d", c.TempDir, time.Now().UnixNano())
	f, err := os.Create(req.path)
	if err != nil {
		return nil, err
	}

	_, err = io.CopyN(f, r, int64(req.Size))
	f.Close()

	if err != nil {
		return nil, err
	}

	return req, nil

}

// ReceiveResponse receives the Server's response. A maxSize parameter must be supplied.
// If the image in the response is larger than maxSize bytes, an error is returned.
func ReceiveResponse(r io.Reader, w io.Writer, maxSize uint32) (id uint64, err error) {
	var buf [12]byte
	if _, err = r.Read(buf[:]); err != nil {
		return
	}

	id = binary.BigEndian.Uint64(buf[:])
	sz := int64(binary.BigEndian.Uint32(buf[8:]))

	if sz > int64(maxSize) {
		return 0, fmt.Errorf("request payload too large (%d bytes)", sz)
	}

	wr, _ := io.CopyN(w, r, sz)
	if wr != int64(sz) {
		return 0, fmt.Errorf("too short")
	}

	return
}

// Send sends the response to the io.Writer.
func (r *Response) Send(w io.Writer) error {
	if len(r.Data) > math.MaxInt32 {
		panic("Response Data too large")
	}

	var buf [12]byte
	binary.BigEndian.PutUint64(buf[:], r.ID)
	binary.BigEndian.PutUint32(buf[8:], uint32(len(r.Data)))

	if _, err := w.Write(buf[:]); err != nil {
		return err
	}

	if _, err := w.Write(r.Data); err != nil {
		return err
	}

	return nil
}
