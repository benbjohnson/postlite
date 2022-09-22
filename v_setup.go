package postlite

import (
	"context"
	"database/sql"
	"fmt"
)

func setupDBServer(ctx context.Context, conn *sql.DB) {
	var setupQueries = []string{
		"CREATE SCHEMA IF NOT EXISTS pg_catalog",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_namespace(oid INT, nspname VARCHAR(128), nspowner INT, nspacl VARCHAR(128)) ON COMMIT PRESERVE ROWS",
		"INSERT INTO pg_catalog.pg_namespace VALUES(99,'pg_toast',10,''),(11,'pg_catalog',10,''), (2200,'public',10,'')",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_description(objoid INT, clasoid INT, objsubid INT, description VARCHAR(128)) ON COMMIT PRESERVE ROWS",
		"INSERT INTO pg_catalog.pg_description VALUES(11,2615,0,'system catalog schema'),(99,2615,0,'reserved schema for TOAST tables'),(2200,2615,0,'standard public schema');",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_database( oid INT, datname VARCHAR(128), datdba INT, encoding INT, datcollate INT, datctype VARCHAR(128), datistemplate INT, datallowconn INT, datconnlimit INT, datlastsysoid INT, datfrozenxid INT, datminmxid INT, dattablespace INT, datacl VARCHAR(128)) ON COMMIT PRESERVE ROWS",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_settings( name VARCHAR(128), setting VARCHAR(128), unit VARCHAR(128), category VARCHAR(128), short_desc VARCHAR(128), extra_desc VARCHAR(128), context VARCHAR(128), vartype VARCHAR(128), source VARCHAR(128), min_val VARCHAR(128), max_val VARCHAR(128), enumvals VARCHAR(128), boot_val VARCHAR(128), reset_val VARCHAR(128), sourcefile VARCHAR(128), sourceline INT, pending_restart INT) ON COMMIT PRESERVE ROWS",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_type( oid INT, typname VARCHAR(128), typnamespace INT, typowner INT, typlen INT, typbyval INT, typtype VARCHAR(128), typcategory VARCHAR(128), typispreferred INT, typisdefined INT, typdelim VARCHAR(128), typrelid INT, typelem INT, typarray INT, typinput VARCHAR(128), typoutput VARCHAR(128), typreceive VARCHAR(128), typsend VARCHAR(128), typmodin VARCHAR(128), typmodout VARCHAR(128), typanalyze VARCHAR(128), typalign VARCHAR(128), typstorage VARCHAR(128), typnotnull INT, typbasetype INT, typtypmod INT, typndims INT, typcollation INT, typdefaultbin VARCHAR(128), typdefault VARCHAR(128), typacl VARCHAR(128)) CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_class( oid INT, relname VARCHAR(128), relnamespace INT, reltype INT, reloftype INT, relowner INT, relam INT, relfilenode INT, reltablespace INT, relpages INT, reltuples REAL, relallvisible INT, reltoastrelid INT, relhasindex INT, relisshared INT, relpersistence VARCHAR(128), relkind VARCHAR(128), relnatts INT, relchecks INT, relhasrules INT, relhastriggers INT, relhassubclass INT, relrowsecurity INT, relforcerowsecurity INT, relispopulated INT, relreplident VARCHAR(128), relispartition INT, relrewrite INT, relfrozenxid INT, relminmxid INT, relacl VARCHAR(128), reloptions VARCHAR(128), relpartbound VARCHAR(128)) ON COMMIT PRESERVE ROWSON COMMIT PRESERVE ROWS",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_class( oid INT, relname VARCHAR(128), relnamespace INT, reltype INT, reloftype INT, relowner INT, relam INT, relfilenode INT, reltablespace INT, relpages INT, reltuples REAL, relallvisible INT, reltoastrelid INT, relhasindex INT, relisshared INT, relpersistence VARCHAR(128), relkind VARCHAR(128), relnatts INT, relchecks INT, relhasrules INT, relhastriggers INT, relhassubclass INT, relrowsecurity INT, relforcerowsecurity INT, relispopulated INT, relreplident VARCHAR(128), relispartition INT, relrewrite INT, relfrozenxid INT, relminmxid INT, relacl VARCHAR(128), reloptions VARCHAR(128), relpartbound VARCHAR(128)) ON COMMIT PRESERVE ROWS",
		"CREATE GLOBAL TEMPORARY TABLE IF NOT EXISTS pg_catalog.pg_range( rngtypid INT, rngsubtype INT, rngmultitypid INT, rngcollation INT, rngsubopc INT, rngcanonical VARCHAR(128), rngsubdiff VARCHAR(128)) ON COMMIT PRESERVE ROWS"}

	for _, query := range setupQueries {
		fmt.Printf("Execute query %s", query)
		conn.QueryContext(ctx, query)
	}
}
