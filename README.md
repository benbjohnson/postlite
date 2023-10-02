Postlite ![Status](https://img.shields.io/badge/status-unmaintained-yellow)
========

Postlite is a network proxy to allow access to remote SQLite databases over the
Postgres wire protocol. This allows GUI tools to be used on remote SQLite
databases which can make administration easier.

The proxy works by translating Postgres frontend wire messages into SQLite
transactions and converting results back into Postgres response wire messages.
Many Postgres clients also inspect the `pg_catalog` to determine system
information so Postlite mirrors this catalog by using an attached in-memory
database with virtual tables. The proxy also performs minor rewriting on these
system queries to convert them to usable SQLite syntax.

_Note: This software was a proof of concept of wrapping SQLite with the Postgres
wire protocol. It is no longer maintained. You're welcome to fork this project if
you're interested in continuing development._


## Usage

To use Postlite, execute the command with the directory that contains your
SQLite databases:

```sh
$ postlite -data-dir /data
```

On another machine, you can connect via the regular Postgres port of 5432:

```sh
$ psql --host HOSTNAME my.db
```

This will connect you to a SQLite database at the path `/data/my.db`.


## Development

Postlite uses virtual tables to simulate the `pg_catalog` so you will need to
enable the `vtable` tag when building:

```sh
$ go install -tags vtable ./cmd/postlite
```

