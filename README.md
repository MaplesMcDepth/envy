# envy

![CI](https://github.com/MaplesMcDepth/envy/actions/workflows/ci.yml/badge.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)


Environment variable manager — set, get, load, and check env vars.

## Install

```bash
go install github.com/MaplesMcDepth/envy/cmd/envy@latest
```

## Commands

### `list` — Show all env vars
```bash
envy list              # All env vars
envy list -s           # Sorted
envy list -e           # Include empty values
```

### `get` — Get a single value
```bash
envy get HOME
envy get DATABASE_URL
```

### `set` — Persist a value to `~/.env`
```bash
envy set API_KEY sk-abc123
envy set DATABASE_URL postgres://localhost/mydb
```

### `unset` — Remove from `~/.env`
```bash
envy unset API_KEY
```

### `load` — Output `export` statements for sourcing
```bash
envy load .env.local
envy load .env.local > /tmp/exports.sh && source /tmp/exports.sh
```

### `check` — See which `~/.env` vars are set in current shell
```bash
envy check
```

## Options

| Flag | Description |
|------|-------------|
| `-f string` | Env file path (default `~/.env`) |
| `-e` | Include empty values |
| `-s` | Sort output |
