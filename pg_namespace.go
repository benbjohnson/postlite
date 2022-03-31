package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgNamespaceModule struct{}

func (m *pgNamespaceModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid      INT,
			nspname  TEXT,
			nspowner INTEGER,
			nspacl   TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgNamespaceTable{}, nil
}

func (m *pgNamespaceModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgNamespaceModule) DestroyModule() {}

type pgNamespaceTable struct{}

func (t *pgNamespaceTable) Open() (sqlite3.VTabCursor, error) {
	return &pgNamespaceCursor{}, nil
}

func (t *pgNamespaceTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgNamespaceTable) Disconnect() error { return nil }
func (t *pgNamespaceTable) Destroy() error    { return nil }

type pgNamespaceCursor struct {
	index int
}

func (c *pgNamespaceCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgNamespaces[c.index].oid)
	case 1:
		sctx.ResultText(pgNamespaces[c.index].nspname)
	case 2:
		sctx.ResultInt(pgNamespaces[c.index].nspowner)
	case 3:
		sctx.ResultText(pgNamespaces[c.index].nspacl)
	}
	return nil
}

func (c *pgNamespaceCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgNamespaceCursor) Next() error {
	c.index++
	return nil
}

func (c *pgNamespaceCursor) EOF() bool {
	return c.index >= len(pgNamespaces)
}

func (c *pgNamespaceCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgNamespaceCursor) Close() error {
	return nil
}

type pgNamespace struct {
	oid      int
	nspname  string
	nspowner int
	nspacl   string
}

var pgNamespaces = []pgNamespace{
	{99, "pg_toast", 10, ""},
	{11, "pg_catalog", 10, ""},
	{2200, "public", 10, ""},
	{13427, "information_schema", 10, ""},
}
