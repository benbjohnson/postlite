package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgDatabaseModule struct{}

func (m *pgDatabaseModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid           INTEGER,
			datname       TEXT,
			datdba        INTEGER,
			encoding      INTEGER,
			datcollate    TEXT,
			datctype      TEXT,
			datistemplate INTEGER,
			datallowconn  INTEGER,
			datconnlimit  INTEGER,
			datlastsysoid INTEGER,
			datfrozenxid  INTEGER,
			datminmxid    INTEGER,
			dattablespace INTEGER,
			datacl        TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgDatabaseTable{}, nil
}

func (m *pgDatabaseModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgDatabaseModule) DestroyModule() {}

type pgDatabaseTable struct{}

func (t *pgDatabaseTable) Open() (sqlite3.VTabCursor, error) {
	return &pgDatabaseCursor{}, nil
}

func (t *pgDatabaseTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgDatabaseTable) Disconnect() error { return nil }
func (t *pgDatabaseTable) Destroy() error    { return nil }

type pgDatabaseCursor struct {
	index int
}

func (c *pgDatabaseCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgDatabases[c.index].oid)
	case 1:
		sctx.ResultText(pgDatabases[c.index].datname)
	case 2:
		sctx.ResultInt(pgDatabases[c.index].datdba)
	case 3:
		sctx.ResultInt(pgDatabases[c.index].encoding)
	case 4:
		sctx.ResultText(pgDatabases[c.index].datcollate)
	case 5:
		sctx.ResultText(pgDatabases[c.index].datctype)
	case 6:
		sctx.ResultInt(pgDatabases[c.index].datistemplate)
	case 7:
		sctx.ResultInt(pgDatabases[c.index].datallowconn)
	case 8:
		sctx.ResultInt(pgDatabases[c.index].datconnlimit)
	case 9:
		sctx.ResultInt(pgDatabases[c.index].datlastsysoid)
	case 10:
		sctx.ResultInt(pgDatabases[c.index].datfrozenxid)
	case 11:
		sctx.ResultInt(pgDatabases[c.index].datminmxid)
	case 12:
		sctx.ResultInt(pgDatabases[c.index].dattablespace)
	case 13:
		sctx.ResultText(pgDatabases[c.index].datacl)
	}
	return nil
}

func (c *pgDatabaseCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgDatabaseCursor) Next() error {
	c.index++
	return nil
}

func (c *pgDatabaseCursor) EOF() bool {
	return c.index >= len(pgDatabases)
}

func (c *pgDatabaseCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgDatabaseCursor) Close() error {
	return nil
}

type pgDatabase struct {
	oid           int
	datname       string
	datdba        int
	encoding      int
	datcollate    string
	datctype      string
	datistemplate int
	datallowconn  int
	datconnlimit  int
	datlastsysoid int
	datfrozenxid  int
	datminmxid    int
	dattablespace int
	datacl        string
}

var pgDatabases = []pgDatabase{}
