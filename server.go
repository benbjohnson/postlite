package postlite

import (
	"context"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	ln net.Listener

	g      errgroup.Group
	ctx    context.Context
	cancel func()

	// Bind address to listen to Postgres wire protocol.
	Addr string
}

func NewServer() *Server {
	s := &Server{}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Open() (err error) {
	s.ln, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.g.Go(s.serve)
	return nil
}

func (s *Server) Close() (err error) {
	if s.ln != nil {
		if e := s.ln.Close(); err == nil {
			err = e
		}
	}
	s.cancel()

	if err := s.g.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Server) serve() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		c := NewConn(conn)

		log.Println("connection accepted: ", c.conn.RemoteAddr())

		s.g.Go(func() error {
			defer c.Close()

			if err := c.Serve(s.ctx); err != nil && s.ctx.Err() == nil {
				log.Println("connection error, closing: %s", err)
				return nil
			}

			log.Printf("connection closed: %s", c.conn.RemoteAddr())
			return nil
		})
	}
}
