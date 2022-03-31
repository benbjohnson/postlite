package postlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

// Postgres settings.
const (
	ServerVersion = "13.0.0"
)

func init() {
	sql.Register("postlite-sqlite3", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("current_catalog", currentCatalog, true); err != nil {
				return fmt.Errorf("cannot register current_catalog() function")
			}
			if err := conn.RegisterFunc("current_schema", currentSchema, true); err != nil {
				return fmt.Errorf("cannot register current_schema() function")
			}
			if err := conn.RegisterFunc("current_user", currentUser, true); err != nil {
				return fmt.Errorf("cannot register current_schema() function")
			}
			if err := conn.RegisterFunc("session_user", sessionUser, true); err != nil {
				return fmt.Errorf("cannot register session_user() function")
			}
			if err := conn.RegisterFunc("user", user, true); err != nil {
				return fmt.Errorf("cannot register user() function")
			}
			if err := conn.RegisterFunc("show", show, true); err != nil {
				return fmt.Errorf("cannot register show() function")
			}
			if err := conn.RegisterFunc("format_type", formatType, true); err != nil {
				return fmt.Errorf("cannot register format_type() function")
			}
			if err := conn.RegisterFunc("version", version, true); err != nil {
				return fmt.Errorf("cannot register version() function")
			}

			if err := conn.CreateModule("pg_namespace_module", &pgNamespaceModule{}); err != nil {
				return fmt.Errorf("cannot register pg_namespace module")
			}
			if err := conn.CreateModule("pg_description_module", &pgDescriptionModule{}); err != nil {
				return fmt.Errorf("cannot register pg_description module")
			}
			if err := conn.CreateModule("pg_database_module", &pgDatabaseModule{}); err != nil {
				return fmt.Errorf("cannot register pg_database module")
			}
			if err := conn.CreateModule("pg_settings_module", &pgSettingsModule{}); err != nil {
				return fmt.Errorf("cannot register pg_settings module")
			}
			if err := conn.CreateModule("pg_type_module", &pgTypeModule{}); err != nil {
				return fmt.Errorf("cannot register pg_type module")
			}
			if err := conn.CreateModule("pg_class_module", &pgClassModule{}); err != nil {
				return fmt.Errorf("cannot register pg_class module")
			}
			return nil
		},
	})
}

func currentCatalog() string { return "public" }
func currentSchema() string  { return "public" }

func currentUser() string { return "sqlite3" }
func sessionUser() string { return "sqlite3" }
func user() string        { return "sqlite3" }

func version() string { return "postlite v0.0.0" }

func formatType(type_oid, typemod string) string { return "" }

func show(name string) string {
	println("dbg/show", name)
	return ""
}

type Server struct {
	mu    sync.Mutex
	ln    net.Listener
	conns map[*Conn]struct{}

	g      errgroup.Group
	ctx    context.Context
	cancel func()

	// Bind address to listen to Postgres wire protocol.
	Addr string

	// Directory that holds SQLite databases.
	DataDir string
}

func NewServer() *Server {
	s := &Server{
		conns: make(map[*Conn]struct{}),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Open() (err error) {
	// Ensure data directory exists.
	if _, err := os.Stat(s.DataDir); err != nil {
		return err
	}

	s.ln, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.g.Go(func() error {
		if err := s.serve(); s.ctx.Err() != nil {
			return err // return error unless context canceled
		}
		return nil
	})
	return nil
}

func (s *Server) Close() (err error) {
	if s.ln != nil {
		if e := s.ln.Close(); err == nil {
			err = e
		}
	}
	s.cancel()

	// Track and close all open connections.
	if e := s.CloseClientConnections(); err == nil {
		err = e
	}

	if err := s.g.Wait(); err != nil {
		return err
	}
	return err
}

// CloseClientConnections disconnects all Postgres connections.
func (s *Server) CloseClientConnections() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.conns {
		if e := conn.Close(); err == nil {
			err = e
		}
	}

	s.conns = make(map[*Conn]struct{})

	return err
}

// CloseClientConnection disconnects a Postgres connections.
func (s *Server) CloseClientConnection(conn *Conn) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, conn)
	return conn.Close()
}

func (s *Server) serve() error {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return err
		}
		conn := newConn(c)

		// Track live connections.
		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()

		log.Println("connection accepted: ", conn.RemoteAddr())

		s.g.Go(func() error {
			defer s.CloseClientConnection(conn)

			if err := s.serveConn(s.ctx, conn); err != nil && s.ctx.Err() == nil {
				log.Printf("connection error, closing: %s", err)
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

		log.Printf("[recv] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Query:
			if err := s.handleQueryMessage(ctx, c, msg); err != nil {
				return fmt.Errorf("query message: %w", err)
			}

		case *pgproto3.Parse:
			if err := s.handleParseMessage(ctx, c, msg); err != nil {
				return fmt.Errorf("parse message: %w", err)
			}

		case *pgproto3.Sync: // ignore
			continue

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

func (s *Server) handleStartupMessage(ctx context.Context, c *Conn, msg *pgproto3.StartupMessage) (err error) {
	log.Printf("received startup message: %#v", msg)

	// Validate
	name := getParameter(msg.Parameters, "database")
	if name == "" {
		return writeMessages(c, &pgproto3.ErrorResponse{Message: "database required"})
	} else if strings.Contains(name, "..") {
		return writeMessages(c, &pgproto3.ErrorResponse{Message: "invalid database name"})
	}

	// Open SQL database & attach to the connection.
	if c.db, err = sql.Open("postlite-sqlite3", filepath.Join(s.DataDir, name)); err != nil {
		return err
	}

	// Register virtual tables to imitate postgres.
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_namespace USING pg_namespace_module (oid, nspname, nspowner, nspacl)"); err != nil {
		return fmt.Errorf("create pg_namespace: %w", err)
	}
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_description USING pg_description_module (objoid, classoid, objsubid, description)"); err != nil {
		return fmt.Errorf("create pg_description: %w", err)
	}
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_database USING pg_database_module (oid, datname, datdba, encoding, datcollate, datctype, datistemplate, datallowconn, datconnlimit, datlastsysoid, datfrozenxid, datminmxid, dattablespace, datacl)"); err != nil {
		return fmt.Errorf("create pg_database: %w", err)
	}
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_settings USING pg_settings_module (name, setting, unit, category, short_desc, extra_desc, context, vartype, source, min_val, max_val, enumvals, boot_val, reset_val, sourcefile, sourceline, pending_restart)"); err != nil {
		return fmt.Errorf("create pg_settings: %w", err)
	}
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_type USING pg_type_module (oid, typname, typnamespace, typowner, typlen, typbyval, typtype, typcategory, typispreferred, typisdefined, typdelim, typrelid, typelem, typarray, typinput, typoutput, typreceive, typsend, typmodin, typmodout, typanalyze, typalign, typstorage, typnotnull, typbasetype, typtypmod, typndims, typcollation, typdefaultbin, typdefault, typacl)"); err != nil {
		return fmt.Errorf("create pg_type: %w", err)
	}
	if _, err := c.db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_class USING pg_class_module (oid, relname, relnamespace, reltype, reloftype, relowner, relam, relfilenode, reltablespace, relpages, reltuples, relallvisible, reltoastrelid, relhasindex, relisshared, relpersistence, relkind, relnatts, relchecks, relhasrules, relhastriggers, relhassubclass, relrowsecurity, relforcerowsecurity, relispopulated, relreplident, relispartition, relrewrite, relfrozenxid, relminmxid, relacl, reloptions, relpartbound)"); err != nil {
		return fmt.Errorf("create pg_class: %w", err)
	}

	return writeMessages(c,
		&pgproto3.AuthenticationOk{},
		&pgproto3.ParameterStatus{Name: "server_version", Value: ServerVersion},
		&pgproto3.ReadyForQuery{TxStatus: 'I'},
	)
}

func (s *Server) handleSSLRequestMessage(ctx context.Context, c *Conn, msg *pgproto3.SSLRequest) error {
	log.Printf("received ssl request message: %#v", msg)
	if _, err := c.Write([]byte("N")); err != nil {
		return err
	}
	return s.serveConnStartup(ctx, c)
}

func (s *Server) handleQueryMessage(ctx context.Context, c *Conn, msg *pgproto3.Query) error {
	log.Printf("received query: %q", msg.String)

	// Execute query against database.
	rows, err := c.db.QueryContext(ctx, msg.String)
	if err != nil {
		return writeMessages(c,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}
	defer rows.Close()

	// Encode column header.
	cols, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}
	buf := toRowDescription(cols).Encode(nil)

	// Iterate over each row and encode it to the wire protocol.
	for rows.Next() {
		row, err := scanRow(rows, cols)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		buf = row.Encode(buf)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}

	// Mark command complete and ready for next query.
	buf = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
	buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

	_, err = c.Write(buf)
	return err
}

func toRowDescription(cols []*sql.ColumnType) *pgproto3.RowDescription {
	var desc pgproto3.RowDescription
	for _, col := range cols {
		desc.Fields = append(desc.Fields, pgproto3.FieldDescription{
			Name:                 []byte(col.Name()),
			TableOID:             0,
			TableAttributeNumber: 0,
			DataTypeOID:          pgtype.TextOID,
			DataTypeSize:         -1,
			TypeModifier:         -1,
			Format:               0,
		})
	}
	return &desc
}

func scanRow(rows *sql.Rows, cols []*sql.ColumnType) (*pgproto3.DataRow, error) {
	refs := make([]interface{}, len(cols))
	values := make([]interface{}, len(cols))
	for i := range refs {
		refs[i] = &values[i]
	}

	// Scan from SQLite database.
	if err := rows.Scan(refs...); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Convert to TEXT values to return over Postgres wire protocol.
	row := pgproto3.DataRow{Values: make([][]byte, len(values))}
	for i := range values {
		row.Values[i] = []byte(fmt.Sprint(values[i]))
	}
	return &row, nil
}

func (s *Server) handleParseMessage(ctx context.Context, c *Conn, pmsg *pgproto3.Parse) error {
	// Rewrite system-information queries so they're tolerable by SQLite.
	query := rewriteQuery(pmsg.Query)

	if pmsg.Query != query {
		log.Printf("query rewrite: %s", query)
	}

	// Prepare the query.
	stmt, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	var rows *sql.Rows
	var cols []*sql.ColumnType
	var binds []interface{}
	exec := func() (err error) {
		if rows != nil {
			return nil
		}
		if rows, err = stmt.QueryContext(ctx, binds...); err != nil {
			return fmt.Errorf("query: %w", err)
		}
		if cols, err = rows.ColumnTypes(); err != nil {
			return fmt.Errorf("column types: %w", err)
		}
		return nil
	}

	// LOOP:
	for {
		msg, err := c.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message during parse: %w", err)
		}

		log.Printf("[recv(p)] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Bind:
			binds = make([]interface{}, len(msg.Parameters))
			for i := range msg.Parameters {
				binds[i] = string(msg.Parameters[i])
			}

		case *pgproto3.Describe:
			if err := exec(); err != nil {
				return fmt.Errorf("exec: %w", err)
			}
			if _, err := c.Write(toRowDescription(cols).Encode(nil)); err != nil {
				return err
			}

		case *pgproto3.Execute:
			// TODO: Send pgproto3.ParseComplete?
			if err := exec(); err != nil {
				return fmt.Errorf("exec: %w", err)
			}

			var buf []byte
			for rows.Next() {
				row, err := scanRow(rows, cols)
				if err != nil {
					return fmt.Errorf("scan: %w", err)
				}
				buf = row.Encode(buf)
			}
			if err := rows.Err(); err != nil {
				return fmt.Errorf("rows: %w", err)
			}

			// Mark command complete and ready for next query.
			buf = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
			buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
			_, err := c.Write(buf)
			return err

		default:
			return fmt.Errorf("unexpected message type during parse: %#v", msg)
		}
	}
}

func (s *Server) execSetQuery(ctx context.Context, c *Conn, query string) error {
	buf := (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(nil)
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

func getParameter(m map[string]string, k string) string {
	if m == nil {
		return ""
	}
	return m[k]
}

// writeMessages writes all messages to a single buffer before sending.
func writeMessages(w io.Writer, msgs ...pgproto3.Message) error {
	var buf []byte
	for _, msg := range msgs {
		buf = msg.Encode(buf)
	}
	_, err := w.Write(buf)
	return err
}

func rewriteQuery(q string) string {
	// Ignore SET queries by rewriting them to empty resultsets.
	if strings.HasPrefix(q, "SET ") {
		return `SELECT 'SET'`
	}

	// Ignore this god forsaken query for pulling keywords.
	if strings.Contains(q, `select string_agg(word, ',') from pg_catalog.pg_get_keywords()`) {
		return `SELECT '' AS "string_agg" WHERE 1 = 2`
	}

	// Rewrite system information variables so they are functions so we can inject them.
	// https://www.postgresql.org/docs/9.1/functions-info.html
	q = systemFunctionRegex.ReplaceAllString(q, "$1()$2")

	// Rewrite double-colon casting by simply removing it.
	// https://www.postgresql.org/docs/7.3/sql-expressions.html#SQL-SYNTAX-TYPE-CASTS
	q = castRegex.ReplaceAllString(q, "")

	// Remove references to the pg_catalog.
	q = pgCatalogRegex.ReplaceAllString(q, "")

	// Rewrite "SHOW" commands into function calls.
	q = showRegex.ReplaceAllString(q, "SELECT show('$1')")

	return q
}

var (
	systemFunctionRegex = regexp.MustCompile(`\b(current_catalog|current_schema|current_user|session_user|user)\b([^\(]|$)`)

	castRegex = regexp.MustCompile(`::(regclass)`)

	pgCatalogRegex = regexp.MustCompile(`\bpg_catalog\.`)

	showRegex = regexp.MustCompile(`^SHOW (\w+)`)
)
