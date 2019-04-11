package thmb

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

// HandleFunc is the signature of the handler function of the server.
type HandleFunc func(r *Request, res *Response) error

// Server is a thumbnailer server that listens for requests and resizes images.
type Server struct {
	c     Config
	ln    net.Listener
	hndlr HandleFunc

	id uint64

	reqCh chan *Request
	resCh chan *Response
}

func defaultHandler(r *Request, res *Response) error {
	out := r.path + ".jpg"

	var cmd *exec.Cmd

	if r.Width != 0 || r.Height != 0 {
		sz := fmt.Sprintf("%dx%d!", r.Width, r.Height)
		cmd = exec.Command("convert", r.path, "-resize", sz, out)
	} else {
		// don't resize, just convert to jpg
		cmd = exec.Command("convert", r.path, out)
	}

	tr := time.AfterFunc(4*time.Second, func() {
		log.Println("error: aborted conversion (timeout)")
		cmd.Process.Kill()
	})
	defer tr.Stop()

	if err := cmd.Start(); err != nil {
		return err
	}

	err := cmd.Wait()
	if err != nil {
		return err
	}

	defer os.Remove(out)
	res.Data, err = ioutil.ReadFile(out)
	return err
}

// NewServer creates a thumbnailer server.
func NewServer(c *Config) (s *Server, err error) {
	s = &Server{
		c:     *c,
		reqCh: make(chan *Request, 2),
		resCh: make(chan *Response, 2),
		hndlr: defaultHandler,
	}
	s.ln, err = net.Listen(c.Network, c.Addr)
	return
}

func (s *Server) worker(N int) {
	for req := range s.reqCh {
		res := &Response{
			from: req.from,
		}
		err := s.hndlr(req, res)
		os.Remove(req.path)

		if err != nil {
			req.from.Close()
			continue
		}
		s.resCh <- res
	}
}

// Serve starts the server. You might provide a custom handler or pass nil
// to use the default handler.
func (s *Server) Serve(customHndlr HandleFunc) error {
	if customHndlr != nil {
		s.hndlr = customHndlr
	}

	for i := 0; i < s.c.NumWorkers; i++ {
		go s.worker(i)
	}

	go func() {
		for res := range s.resCh {
			res.Send(res.from)
		}
	}()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println("Accept: ", err)
			return err
		}

		req, err := ReceiveRequest(conn, &s.c)
		if err != nil {
			log.Println("ReceiveRequest: ", err)
			continue
		}
		req.from = conn
		s.id++

		s.reqCh <- req
	}
}
