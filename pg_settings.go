package postlite

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type pgSettingsModule struct{}

func (m *pgSettingsModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			name            TEXT,
			setting         TEXT,
			unit            TEXT,
			category        TEXT,
			short_desc      TEXT,
			extra_desc      TEXT,
			context         TEXT,
			vartype         TEXT,
			source          TEXT,
			min_val         TEXT,
			max_val         TEXT,
			enumvals        TEXT,
			boot_val        TEXT,
			reset_val       TEXT,
			sourcefile      TEXT,
			sourceline      INTEGER,
			pending_restart INTEGER
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &pgSettingsTable{}, nil
}

func (m *pgSettingsModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *pgSettingsModule) DestroyModule() {}

type pgSettingsTable struct{}

func (t *pgSettingsTable) Open() (sqlite3.VTabCursor, error) {
	return &pgSettingsCursor{}, nil
}

func (t *pgSettingsTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *pgSettingsTable) Disconnect() error { return nil }
func (t *pgSettingsTable) Destroy() error    { return nil }

type pgSettingsCursor struct {
	index int
}

func (c *pgSettingsCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultText(pgSettings[c.index].name)
	case 1:
		sctx.ResultText(pgSettings[c.index].setting)
	case 2:
		sctx.ResultText(pgSettings[c.index].unit)
	case 3:
		sctx.ResultText(pgSettings[c.index].category)
	case 4:
		sctx.ResultText(pgSettings[c.index].short_desc)
	case 5:
		sctx.ResultText(pgSettings[c.index].extra_desc)
	case 6:
		sctx.ResultText(pgSettings[c.index].context)
	case 7:
		sctx.ResultText(pgSettings[c.index].vartype)
	case 8:
		sctx.ResultText(pgSettings[c.index].source)
	case 9:
		sctx.ResultText(pgSettings[c.index].min_val)
	case 10:
		sctx.ResultText(pgSettings[c.index].max_val)
	case 11:
		sctx.ResultText(pgSettings[c.index].enumvals)
	case 12:
		sctx.ResultText(pgSettings[c.index].boot_val)
	case 13:
		sctx.ResultText(pgSettings[c.index].reset_val)
	case 14:
		sctx.ResultText(pgSettings[c.index].sourcefile)
	case 15:
		sctx.ResultInt(pgSettings[c.index].sourceline)
	case 16:
		sctx.ResultInt(pgSettings[c.index].pending_restart)
	}
	return nil
}

func (c *pgSettingsCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *pgSettingsCursor) Next() error {
	c.index++
	return nil
}

func (c *pgSettingsCursor) EOF() bool {
	return c.index >= len(pgSettings)
}

func (c *pgSettingsCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *pgSettingsCursor) Close() error {
	return nil
}

type pgSetting struct {
	name            string
	setting         string
	unit            string
	category        string
	short_desc      string
	extra_desc      string
	context         string
	vartype         string
	source          string
	min_val         string
	max_val         string
	enumvals        string
	boot_val        string
	reset_val       string
	sourcefile      string
	sourceline      int
	pending_restart int
}

var pgSettings = []pgSetting{}
