package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgClassModule struct{}

func (m *pgClassModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid                 INTEGER,
			relname             TEXT,
			relnamespace        INTEGER,
			reltype             INTEGER,
			reloftype           INTEGER,
			relowner            INTEGER,
			relam               INTEGER,
			relfilenode         INTEGER,
			reltablespace       INTEGER,
			relpages            INTEGER,
			reltuples           REAL,
			relallvisible       INTEGER,
			reltoastrelid       INTEGER,
			relhasindex         INTEGER,
			relisshared         INTEGER,
			relpersistence      TEXT,
			relkind             TEXT,
			relnatts            INTEGER,
			relchecks           INTEGER,
			relhasrules         INTEGER,
			relhastriggers      INTEGER,
			relhassubclass      INTEGER,
			relrowsecurity      INTEGER,
			relforcerowsecurity INTEGER,
			relispopulated      INTEGER,
			relreplident        TEXT,
			relispartition      INTEGER,
			relrewrite          INTEGER,
			relfrozenxid        INTEGER,
			relminmxid          INTEGER,
			relacl              TEXT,
			reloptions          TEXT,
			relpartbound        TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgClassTable{}, nil
}

func (m *pgClassModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgClassModule) DestroyModule() {}

type pgClassTable struct{}

func (t *pgClassTable) Open() (sqlite3.VTabCursor, error) {
	return &pgClassCursor{}, nil
}

func (t *pgClassTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgClassTable) Disconnect() error { return nil }
func (t *pgClassTable) Destroy() error    { return nil }

type pgClassCursor struct {
	index int
}

func (c *pgClassCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgClasses[c.index].oid)
	case 1:
		sctx.ResultText(pgClasses[c.index].relname)
	case 2:
		sctx.ResultInt(pgClasses[c.index].relnamespace)
	case 3:
		sctx.ResultInt(pgClasses[c.index].reltype)
	case 4:
		sctx.ResultInt(pgClasses[c.index].reloftype)
	case 5:
		sctx.ResultInt(pgClasses[c.index].relowner)
	case 6:
		sctx.ResultInt(pgClasses[c.index].relam)
	case 7:
		sctx.ResultInt(pgClasses[c.index].relfilenode)
	case 8:
		sctx.ResultInt(pgClasses[c.index].reltablespace)
	case 9:
		sctx.ResultInt(pgClasses[c.index].relpages)
	case 10:
		sctx.ResultDouble(pgClasses[c.index].reltuples)
	case 11:
		sctx.ResultInt(pgClasses[c.index].relallvisible)
	case 12:
		sctx.ResultInt(pgClasses[c.index].reltoastrelid)
	case 13:
		sctx.ResultInt(pgClasses[c.index].relhasindex)
	case 14:
		sctx.ResultInt(pgClasses[c.index].relisshared)
	case 15:
		sctx.ResultText(pgClasses[c.index].relpersistence)
	case 16:
		sctx.ResultText(pgClasses[c.index].relkind)
	case 17:
		sctx.ResultInt(pgClasses[c.index].relnatts)
	case 18:
		sctx.ResultInt(pgClasses[c.index].relchecks)
	case 19:
		sctx.ResultInt(pgClasses[c.index].relhasrules)
	case 20:
		sctx.ResultInt(pgClasses[c.index].relhastriggers)
	case 21:
		sctx.ResultInt(pgClasses[c.index].relhassubclass)
	case 22:
		sctx.ResultInt(pgClasses[c.index].relrowsecurity)
	case 23:
		sctx.ResultInt(pgClasses[c.index].relforcerowsecurity)
	case 24:
		sctx.ResultInt(pgClasses[c.index].relispopulated)
	case 25:
		sctx.ResultText(pgClasses[c.index].relreplident)
	case 26:
		sctx.ResultInt(pgClasses[c.index].relispartition)
	case 27:
		sctx.ResultInt(pgClasses[c.index].relrewrite)
	case 28:
		sctx.ResultInt(pgClasses[c.index].relfrozenxid)
	case 29:
		sctx.ResultInt(pgClasses[c.index].relminmxid)
	case 30:
		sctx.ResultText(pgClasses[c.index].relacl)
	case 31:
		sctx.ResultText(pgClasses[c.index].reloptions)
	case 32:
		sctx.ResultText(pgClasses[c.index].relpartbound)
	}
	return nil
}

func (c *pgClassCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgClassCursor) Next() error {
	c.index++
	return nil
}

func (c *pgClassCursor) EOF() bool {
	return c.index >= len(pgClasses)
}

func (c *pgClassCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgClassCursor) Close() error {
	return nil
}

type pgClass struct {
	oid                 int
	relname             string
	relnamespace        int
	reltype             int
	reloftype           int
	relowner            int
	relam               int
	relfilenode         int
	reltablespace       int
	relpages            int
	reltuples           float64
	relallvisible       int
	reltoastrelid       int
	relhasindex         int
	relisshared         int
	relpersistence      string
	relkind             string
	relnatts            int
	relchecks           int
	relhasrules         int
	relhastriggers      int
	relhassubclass      int
	relrowsecurity      int
	relforcerowsecurity int
	relispopulated      int
	relreplident        string
	relispartition      int
	relrewrite          int
	relfrozenxid        int
	relminmxid          int
	relacl              string
	reloptions          string
	relpartbound        string
}

var pgClasses = []pgClass{}
