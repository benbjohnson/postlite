package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgTypeModule struct{}

func (m *pgTypeModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid            INTEGER,
			typname        TEXT,
			typnamespace   INTEGER,
			typowner       INTEGER,
			typlen         INTEGER,
			typbyval       INTEGER,
			typtype        TEXT,
			typcategory    TEXT,
			typispreferred INTEGER,
			typisdefined   INTEGER,
			typdelim       TEXT,
			typrelid       INTEGER,
			typelem        INTEGER,
			typarray       INTEGER,
			typinput       TEXT,
			typoutput      TEXT,
			typreceive     TEXT,
			typsend        TEXT,
			typmodin       TEXT,
			typmodout      TEXT,
			typanalyze     TEXT,
			typalign       TEXT,
			typstorage     TEXT,
			typnotnull     INTEGER,
			typbasetype    INTEGER,
			typtypmod      INTEGER,
			typndims       INTEGER,
			typcollation   INTEGER,
			typdefaultbin  TEXT,
			typdefault     TEXT,
			typacl         TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgTypeTable{}, nil
}

func (m *pgTypeModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgTypeModule) DestroyModule() {}

type pgTypeTable struct{}

func (t *pgTypeTable) Open() (sqlite3.VTabCursor, error) {
	return &pgTypeCursor{}, nil
}

func (t *pgTypeTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgTypeTable) Disconnect() error { return nil }
func (t *pgTypeTable) Destroy() error    { return nil }

type pgTypeCursor struct {
	index int
}

func (c *pgTypeCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgTypes[c.index].oid)
	case 1:
		sctx.ResultText(pgTypes[c.index].typname)
	case 2:
		sctx.ResultInt(pgTypes[c.index].typnamespace)
	case 3:
		sctx.ResultInt(pgTypes[c.index].typowner)
	case 4:
		sctx.ResultInt(pgTypes[c.index].typlen)
	case 5:
		sctx.ResultInt(pgTypes[c.index].typbyval)
	case 6:
		sctx.ResultText(pgTypes[c.index].typtype)
	case 7:
		sctx.ResultText(pgTypes[c.index].typcategory)
	case 8:
		sctx.ResultInt(pgTypes[c.index].typispreferred)
	case 9:
		sctx.ResultInt(pgTypes[c.index].typisdefined)
	case 10:
		sctx.ResultText(pgTypes[c.index].typdelim)
	case 11:
		sctx.ResultInt(pgTypes[c.index].typrelid)
	case 12:
		sctx.ResultInt(pgTypes[c.index].typelem)
	case 13:
		sctx.ResultInt(pgTypes[c.index].typarray)
	case 14:
		sctx.ResultText(pgTypes[c.index].typinput)
	case 15:
		sctx.ResultText(pgTypes[c.index].typoutput)
	case 16:
		sctx.ResultText(pgTypes[c.index].typreceive)
	case 17:
		sctx.ResultText(pgTypes[c.index].typsend)
	case 18:
		sctx.ResultText(pgTypes[c.index].typmodin)
	case 19:
		sctx.ResultText(pgTypes[c.index].typmodout)
	case 20:
		sctx.ResultText(pgTypes[c.index].typanalyze)
	case 21:
		sctx.ResultText(pgTypes[c.index].typalign)
	case 22:
		sctx.ResultText(pgTypes[c.index].typstorage)
	case 23:
		sctx.ResultInt(pgTypes[c.index].typnotnull)
	case 24:
		sctx.ResultInt(pgTypes[c.index].typbasetype)
	case 25:
		sctx.ResultInt(pgTypes[c.index].typtypmod)
	case 26:
		sctx.ResultInt(pgTypes[c.index].typndims)
	case 27:
		sctx.ResultInt(pgTypes[c.index].typcollation)
	case 28:
		sctx.ResultText(pgTypes[c.index].typdefaultbin)
	case 29:
		sctx.ResultText(pgTypes[c.index].typdefault)
	case 30:
		sctx.ResultText(pgTypes[c.index].typacl)
	}
	return nil
}

func (c *pgTypeCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgTypeCursor) Next() error {
	c.index++
	return nil
}

func (c *pgTypeCursor) EOF() bool {
	return c.index >= len(pgTypes)
}

func (c *pgTypeCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgTypeCursor) Close() error {
	return nil
}

type pgType struct {
	oid            int
	typname        string
	typnamespace   int
	typowner       int
	typlen         int
	typbyval       int
	typtype        string
	typcategory    string
	typispreferred int
	typisdefined   int
	typdelim       string
	typrelid       int
	typelem        int
	typarray       int
	typinput       string
	typoutput      string
	typreceive     string
	typsend        string
	typmodin       string
	typmodout      string
	typanalyze     string
	typalign       string
	typstorage     string
	typnotnull     int
	typbasetype    int
	typtypmod      int
	typndims       int
	typcollation   int
	typdefaultbin  string
	typdefault     string
	typacl         string
}

var pgTypes = []pgType{}
