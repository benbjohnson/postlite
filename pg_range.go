package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgRangeModule struct{}

func (m *pgRangeModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			rngtypid 	INTEGER,
			rngsubtype	INTEGER,
			rngmultitypid	INTEGER,
			rngcollation	INTEGER,
			rngsubopc	INTEGER,
			rngcanonical	TEXT,
			rngsubdiff	TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgNamespaceTable{}, nil
}

func (m *pgRangeModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgRangeModule) DestroyModule() {}

type pgRangeTable struct{}

func (t *pgRangeTable) Open() (sqlite3.VTabCursor, error) {
	return &pgTypeCursor{}, nil
}

func (t *pgRangeTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgRangeTable) Disconnect() error { return nil }
func (t *pgRangeTable) Destroy() error    { return nil }

type pgRangeCursor struct {
	index int
}

func (c *pgRangeCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgRanges[c.index].rngtypid)
	case 1:
		sctx.ResultInt(pgRanges[c.index].rngsubtype)
	case 2:
		sctx.ResultInt(pgRanges[c.index].rngmultitypid)
	case 3:
		sctx.ResultInt(pgRanges[c.index].rngcollation)
	case 4:
		sctx.ResultInt(pgRanges[c.index].rngsubopc)
	case 5:
		sctx.ResultText(pgRanges[c.index].rngcanonical)
	case 6:
		sctx.ResultText(pgRanges[c.index].rngsubdiff)
	}
	return nil
}

func (c *pgRangeCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgRangeCursor) Next() error {
	c.index++
	return nil
}

func (c *pgRangeCursor) EOF() bool {
	return c.index >= len(pgTypes)
}

func (c *pgRangeCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgRangeCursor) Close() error {
	return nil
}

type pgRange struct {
	rngtypid      int
	rngsubtype    int
	rngmultitypid int
	rngcollation  int
	rngsubopc     int
	rngcanonical  string
	rngsubdiff    string
}

var pgRanges = []pgRange{
	{3904, 23, 4451, 0, 1978, "", ""},
	{3906, 1700, 4532, 0, 3125, "", ""},
	{3908, 1114, 4533, 0, 3128, "", ""},
	{3910, 1184, 4534, 0, 3127, "", ""},
	{3912, 1082, 4535, 0, 3122, "", ""},
	{3926, 20, 4536, 0, 3124, "", ""},
}
