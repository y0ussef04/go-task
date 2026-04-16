# ⚡ Dynamic Database Management System

A mini **phpMyAdmin-like system** built with Go and MySQL. Fully dynamic — no hardcoded schemas. Fully containerized — runs with one command.

## Features

- **Dynamic Schema Engine** — Create databases, tables, and columns at runtime
- **Full CRUD** — Insert, view, update, and delete records for any table
- **Auto-generated Forms** — UI forms are built dynamically from live table schema
- **Clean Architecture** — Handlers → Services → Repository layers
- **Prepared Statements** — No raw SQL concatenation. All user data goes through `?` placeholders
- **Docker Ready** — Single `docker-compose up --build` to run everything

Then open **http://localhost:8081** in your browser.

## CLI Interface

In addition to the web GUI, you can use the interactive CLI tool built directly into the Docker environment.

To start the CLI:
```bash
docker compose run --rm cli
```

**Why use the CLI?**
- **Bulk Operations**: Create sample databases and tables quickly.
- **Direct Access**: Interact with the database from your terminal without opening a browser.
- **Clean Environment**: The CLI connects to the same MySQL container as the GUI.

## Project Structure

```
go-task/
├── main.go                          # Entry point
├── internal/
│   ├── config/config.go             # Env-based configuration
│   ├── models/models.go             # Domain models (Column, Record, APIResponse)
│   ├── repository/mysql_repo.go     # MySQL data access layer
│   ├── service/db_service.go        # Business logic layer
│   └── handler/handler.go           # HTTP handlers + routing
├── static/
│   ├── index.html                   # Single-page UI
│   ├── styles.css                   # Dark-mode CSS
│   └── app.js                       # Frontend logic
├── Dockerfile                       # Multi-stage Go build
└── docker-compose.yml               # App + MySQL orchestration
```

## Architecture

```
Browser  →  Handler (HTTP)  →  Service (Logic)  →  Repository (MySQL)
                                                         ↓
                                                    database/sql
                                                   + prepared stmts
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/databases` | List all databases |
| POST | `/api/databases` | Create database |
| GET | `/api/databases/{db}/tables` | List tables |
| POST | `/api/databases/{db}/tables` | Create table |
| GET | `/api/databases/{db}/tables/{t}/columns` | Get schema |
| POST | `/api/databases/{db}/tables/{t}/columns` | Add column |
| GET | `/api/databases/{db}/tables/{t}/records` | Get records |
| POST | `/api/databases/{db}/tables/{t}/records` | Insert record |
| PUT | `/api/databases/{db}/tables/{t}/records` | Update record |
| DELETE | `/api/databases/{db}/tables/{t}/records` | Delete record |
| POST | `/api/sample` | Create sample RealEstate DB |

## Docker Configuration

| Service | Image | Port |
|---------|-------|------|
| `app` | Custom (Go 1.22 + Alpine) | `8081` |
| `mysql` | mysql:8 | `3306` |

- **MySQL root password**: `secret`
- **Health check**: MySQL must be healthy before the app starts
- **Data persistence**: MySQL data stored in a named volume

## How It Works

1. **Dashboard** — Click a database card to select it
2. **Create Table** — Dynamic column builder with type dropdowns
3. **Insert Record** — Form auto-generates from table schema
4. **Browse Data** — View records in a table, click ✏️ to edit or 🗑️ to delete
5. **Add Column** — Alter any table with a new column at runtime

## Security

- All identifiers (database, table, column names) are validated: alphanumeric + underscore only
- All identifiers are backtick-quoted in SQL
- All data values use prepared statement `?` placeholders
- No raw SQL string concatenation for user data