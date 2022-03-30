Postlite
========

Postlite is a network proxy to allow access to remote SQLite databases over the
Postgres wire protocol.


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

