package postlite

import (
	"context"
	"fmt"
	"net"

	"github.com/jackc/pgproto3/v2"
)

type Conn struct {
	conn    net.Conn
	backend *pgproto3.Backend
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:    conn,
		backend: pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn),
	}
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Serve(ctx context.Context) error {
	if err := c.handleStartup(ctx); err != nil {
		return fmt.Errorf("startup: %w", err)
	}

	for {
		msg, err := c.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message: %w", err)
		}

		switch msg.(type) {
		case *pgproto3.Query:
			response := []byte("foobar")

			buf := (&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
				{
					Name:                 []byte("fortune"),
					TableOID:             0,
					TableAttributeNumber: 0,
					DataTypeOID:          25,
					DataTypeSize:         -1,
					TypeModifier:         -1,
					Format:               0,
				},
			}}).Encode(nil)
			buf = (&pgproto3.DataRow{Values: [][]byte{response}}).Encode(buf)
			buf = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
			buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

			if _, err = c.conn.Write(buf); err != nil {
				return fmt.Errorf("write response: %w", err)
			}
		case *pgproto3.Terminate:
			return nil

		default:
			return fmt.Errorf("unexpected message type: %#v", msg)
		}
	}
}

func (c *Conn) handleStartup(ctx context.Context) error {
	msg, err := c.backend.ReceiveStartupMessage()
	if err != nil {
		return fmt.Errorf("receive startup message: %w", err)
	}

	switch msg.(type) {
	case *pgproto3.StartupMessage:
		buf := (&pgproto3.AuthenticationOk{}).Encode(nil)
		buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

		if _, err := c.conn.Write(buf); err != nil {
			return fmt.Errorf("write auth ok: %w", err)
		}
		return nil

	case *pgproto3.SSLRequest:
		_, err = c.conn.Write([]byte("N"))
		if err != nil {
			return fmt.Errorf("write ssl deny: %w", err)
		}
		return c.handleStartup(ctx)

	default:
		return fmt.Errorf("unexpected startup message: %#v", msg)
	}
}
