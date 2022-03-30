package postlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
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
		c, err := s.ln.Accept()
		if err != nil {
			return err
		}
		conn := newConn(c)

		log.Println("connection accepted: ", conn.RemoteAddr())

		s.g.Go(func() error {
			defer c.Close()

			if err := s.serveConn(s.ctx, conn); err != nil && s.ctx.Err() == nil {
				log.Println("connection error, closing: %s", err)
				return nil
			}

			log.Printf("connection closed: %s", conn.RemoteAddr())
			return nil
		})
	}
}

func (s *Server) serveConn(ctx context.Context, c *Conn) error {
	if err := s.serveConnStartup(ctx, c); err != nil {
		return fmt.Errorf("startup: %w", err)
	}

	for {
		msg, err := c.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message: %w", err)
		}

		switch msg := msg.(type) {
		case *pgproto3.Query:
			if err := s.handleQueryMessage(ctx, c, msg); err != nil {
				return fmt.Errorf("query message: %w", err)
			}
		case *pgproto3.Terminate:
			return nil // exit
		default:
			return fmt.Errorf("unexpected message type: %#v", msg)
		}
	}
}

func (s *Server) serveConnStartup(ctx context.Context, c *Conn) error {
	msg, err := c.backend.ReceiveStartupMessage()
	if err != nil {
		return fmt.Errorf("receive startup message: %w", err)
	}

	switch msg := msg.(type) {
	case *pgproto3.StartupMessage:
		if err := s.handleStartupMessage(ctx, c, msg); err != nil {
			return fmt.Errorf("startup message: %w", err)
		}
		return nil
	case *pgproto3.SSLRequest:
		if err := s.handleSSLRequestMessage(ctx, c, msg); err != nil {
			return fmt.Errorf("ssl request message: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unexpected startup message: %#v", msg)
	}
}

func (s *Server) handleStartupMessage(ctx context.Context, c *Conn, msg *pgproto3.StartupMessage) error {
	log.Printf("received startup message: %#v", msg)

	// TODO: Open SQL database for filepath.Join(s.DataDir, msg.database); attach to conn.db.

	buf := (&pgproto3.AuthenticationOk{}).Encode(nil)
	buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

	_, err := c.Write(buf)
	return err
}

func (s *Server) handleSSLRequestMessage(ctx context.Context, c *Conn, msg *pgproto3.SSLRequest) error {
	if _, err := c.Write([]byte("N")); err != nil {
		return err
	}
	return s.serveConnStartup(ctx, c)
}

func (s *Server) handleQueryMessage(ctx context.Context, c *Conn, msg *pgproto3.Query) error {
	log.Printf("received query: %q", msg.String)

	// TODO: Verify conn.db exists.
	// TODO: Execute msg.String against database.
	// TODO:

	response := []byte("foobar")

	buf := (&pgproto3.RowDescription{
		Fields: []pgproto3.FieldDescription{
			{
				Name:                 []byte("fortune"),
				TableOID:             0,
				TableAttributeNumber: 0,
				DataTypeOID:          pgtype.TextOID,
				DataTypeSize:         -1,
				TypeModifier:         -1,
				Format:               0,
			},
		},
	}).Encode(nil)

	// TODO: Iterate over rows.
	buf = (&pgproto3.DataRow{Values: [][]byte{response}}).Encode(buf)

	buf = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
	buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

	_, err := c.Write(buf)
	return err
}

type Conn struct {
	net.Conn
	backend *pgproto3.Backend
	db      *sql.DB // sqlite database
}

func newConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:    conn,
		backend: pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn),
	}
}
