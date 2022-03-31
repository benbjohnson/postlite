package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgDescriptionModule struct{}

func (m *pgDescriptionModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			objoid      INTEGER,
			classoid    INTEGER,
			objsubid    INTEGER,
			description TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgDescriptionTable{}, nil
}

func (m *pgDescriptionModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgDescriptionModule) DestroyModule() {}

type pgDescriptionTable struct{}

func (t *pgDescriptionTable) Open() (sqlite3.VTabCursor, error) {
	return &pgDescriptionCursor{}, nil
}

func (t *pgDescriptionTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgDescriptionTable) Disconnect() error { return nil }
func (t *pgDescriptionTable) Destroy() error    { return nil }

type pgDescriptionCursor struct {
	index int
}

func (c *pgDescriptionCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgDescriptions[c.index].objoid)
	case 1:
		sctx.ResultInt(pgDescriptions[c.index].classoid)
	case 2:
		sctx.ResultInt(pgDescriptions[c.index].objsubid)
	case 3:
		sctx.ResultText(pgDescriptions[c.index].description)
	}
	return nil
}

func (c *pgDescriptionCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgDescriptionCursor) Next() error {
	c.index++
	return nil
}

func (c *pgDescriptionCursor) EOF() bool {
	return c.index >= len(pgDescriptions)
}

func (c *pgDescriptionCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgDescriptionCursor) Close() error {
	return nil
}

type pgDescription struct {
	objoid      int
	classoid    int
	objsubid    int
	description string
}

var pgDescriptions = []pgDescription{
	{11, 2615, 0, "system catalog schema"},
	{99, 2615, 0, "reserved schema for TOAST tables"},
	{2200, 2615, 0, "standard public schema"},
}
