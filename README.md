Postlite ![Status](https://img.shields.io/badge/status-alpha-yellow)
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

_Note: This software is in alpha. Please report bugs. Postlite doesn't alter
your database unless you issue INSERT, UPDATE, DELETE commands so it's probably
safe. If anything, the Postlite process may die but it shouldn't affect your
database._


## Supported clients

Postgres clients can be quite particular about how they initialize so not all
clients may work. Below are the clients that are currently being tested. If you
would like to see more clients supported or if you're having issues with
existing clients, please [submit an issue][new-issue]!

- [psql](https://www.postgresql.org/docs/current/app-psql.html)
- [DBeaver](https://dbeaver.io/)
- [Postico 2](https://eggerapps.at/postico2/)


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


## Contribution Policy

Postlite is open to code contributions for bug fixes & documentation fixes only.
Features carry a long-term maintenance burden so they will not be accepted at
this time. Please [submit an issue][new-issue] if you have a feature you'd like
to request.

[new-issue]: https://github.com/benbjohnson/postlite/issues/new

