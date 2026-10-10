# todo-ist

A small task manager: a Go REST API backed by SQLite, serving a vanilla HTML/CSS/JS frontend.

## Structure

```
.
├── app/                 # Static frontend (index.html, app.js, styles.css)
├── server/
│   ├── main.go          # HTTP server, basic auth, graceful shutdown
│   ├── api/             # HTTP handlers, validation and tests
│   ├── repository/      # SQL queries for tasks
│   └── database/        # SQLite connection and schema
├── Dockerfile
└── fly.toml             # Fly.io deployment
```

## Requirements

- Go 1.26+
- A C compiler (`gcc`), since `go-sqlite3` requires CGO

## Running locally

```bash
cd server
go run .
```

Open http://localhost:8080. The database file is created on first run.

## Configuration

| Variable          | Default       | Description                                         |
| ----------------- | ------------- | --------------------------------------------------- |
| `PORT`            | `8080`        | HTTP port                                           |
| `DB_PATH`         | `todo-ist.db` | SQLite database file                                |
| `STATIC_DIR`      | `../app`      | Directory served at `/`                             |
| `BASIC_AUTH_USER` | —             | Enables HTTP basic auth when set together with pass |
| `BASIC_AUTH_PASS` | —             | Enables HTTP basic auth when set together with user |

## API

| Method   | Path          | Description    | Success |
| -------- | ------------- | -------------- | ------- |
| `GET`    | `/todos`      | List all tasks | `200`   |
| `GET`    | `/todos/{id}` | Get one task   | `200`   |
| `POST`   | `/todos`      | Create a task  | `201`   |
| `PATCH`  | `/todos/{id}` | Update a task  | `204`   |
| `DELETE` | `/todos/{id}` | Delete a task  | `204`   |

`POST` and `PATCH` take the same body:

```json
{
  "title": "Buy coffee",
  "description": "Optional",
  "status": "pending"
}
```

`title` is required. `status` must be one of `pending`, `in_progress`, `completed` or `cancelled`.

Errors are returned as plain text: `400` for invalid input, `404` when the task does not exist.

```bash
curl -X POST localhost:8080/todos -d '{"title":"Buy coffee","status":"pending"}'
```

## Tests

```bash
cd server
go vet ./...
go test -race ./...
```

## Deployment

CI (`.github/workflows/ci.yml`) runs vet and tests on every push and pull request. Pushes to `master` deploy to Fly.io, with the database stored on the `data` volume mounted at `/data`.

To run the container locally:

```bash
docker build -t todo-ist .
docker run -p 8080:8080 -v todo-data:/data todo-ist
```
