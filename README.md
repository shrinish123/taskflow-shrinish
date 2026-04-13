# TaskFlow

A task management REST API built with Go and PostgreSQL. Users can register, log in, create projects, manage tasks with statuses and priorities, and assign tasks to team members.

## Tech Stack

| Layer    | Technology                                                            |
|----------|-----------------------------------------------------------------------|
| Backend  | Go, chi router, pgx/v5, golang-migrate, JWT (HS256), bcrypt, UUID v7 |
| Database | PostgreSQL 16                                                         |
| Infra    | Docker, Docker Compose                                                |
| Testing  | Postman collection, unit tests (mocks), integration tests (real DB)   |

---

## Architecture Decisions

### Layered Architecture (Handler → Service → Repository)

**Why:** The domain is simple — three entities with straightforward CRUD operations. Hexagonal architecture (ports/adapters) would add indirection without proportional benefit at this scale. A clean three-layer separation keeps the code easy to follow, test, and extend.

**Trade-off:** If the application grew to require complex domain logic (workflow engines, event sourcing), the service layer would need refactoring into richer domain objects. For the current scope, this is the right level of abstraction.

### Interface-Based Dependency Injection

**Why:** Services depend on repository interfaces, not concrete types. This makes the service layer fully testable without a database — unit tests use lightweight mock implementations. The interfaces are defined in the service package (consumer-side), following Go convention.

### Structured Error Handling (AppError)

**Why:** The service layer returns `AppError` values that carry HTTP status codes and user-facing messages. The handler layer calls `WriteError(w, err)` — no switch statements, no error classification. This keeps error semantics in the service where the business logic lives, and keeps handlers as dumb pipes.

```
repository → returns repository.ErrNotFound
service    → mapRepoErr() translates to AppError{404, "project not found"}
handler    → WriteError(w, err) — writes the status + message directly
```

### pgx/v5 without an ORM

**Why:** ORMs in Go (GORM, ent) tend to fight the language's type system and introduce query opacity. With only three tables, raw SQL via `pgx` gives full control, explicit visibility into queries, and better performance. Every query is a single, reviewable SQL string.

**Trade-off:** More boilerplate for CRUD operations. For a larger schema, an ORM or query builder (sqlc) would reduce repetition.

### UUID v7 as Primary Keys

**Why:** UUIDs prevent information leakage (sequential IDs reveal entity counts), allow ID generation before INSERT, and avoid coordination between distributed systems.

We use **UUID v7** (RFC 9562) instead of the more common UUID v4. UUID v4 is fully random, which causes severe B-tree index fragmentation in PostgreSQL — each insert lands at a random leaf page, resulting in ~500x more page splits and ~27% larger indexes compared to sequential IDs. Benchmarks show insert performance degrading by 2-10x at scale.

UUID v7 embeds a millisecond-precision Unix timestamp in the high-order bits, making IDs roughly monotonically increasing. This gives B-tree-friendly insert ordering (new rows append near the end of the index) while retaining global uniqueness and non-guessability. IDs are generated in Go via `github.com/google/uuid.NewV7()` rather than in PostgreSQL, keeping the application in control and avoiding the need for database extensions (PostgreSQL 18 adds native `uuidv7()`, but we target PG 16).

### Migrations Run on Server Startup

**Why:** Simplicity for a docker-compose demo setup. The Go binary runs `golang-migrate` before starting the HTTP server, ensuring the schema is always current. Idempotency is guaranteed by the `schema_migrations` table.

**Trade-off:** In production, migrations should run separately (init container, CI step) to avoid race conditions with multiple replicas. For a single-container demo, startup migrations are the pragmatic choice.

---

## Running Locally

Prerequisites: Docker and Docker Compose installed.

```bash
git clone https://github.com/shrinish123/taskflow-shrinish
cd taskflow-shrinish
cp .env.example .env
docker compose up --build
```

Once both services are healthy, the API server logs will print:

```
=== TaskFlow API server ready ===
API running        url=http://localhost:8080
Health check       url=http://localhost:8080/api/health
Postman collection file=postman_collection.json
Test credentials   email=test@example.com  password=password123
```

| What | URL |
|------|-----|
| **API** | http://localhost:8080 |
| **Health check** | http://localhost:8080/api/health |

The first startup takes a minute to build the Docker image. Subsequent starts are cached.

All endpoints, request bodies, and response examples are documented in the [Postman Collection](postman_collection.json) included in the repo root.

---

## Running Migrations

Migrations run **automatically** on API server startup. No manual steps needed.

If you need to run them manually:

```bash
docker compose exec api ./server  # migrations run before serving
```

The migration files are in `backend/migrations/` and follow the `NNN_description.{up,down}.sql` naming convention.

---

## Test Credentials

Seed data is loaded automatically on first startup:

| Field    | Value             |
|----------|-------------------|
| Email    | test@example.com  |
| Password | password123       |

The seed also creates:
- 1 project ("My First Project")
- 3 tasks with different statuses (todo, in_progress, done)

---

## API Reference

All endpoints, request bodies, and response examples are documented in the **[Postman Collection](postman_collection.json)**.

Import it into Postman — `base_url` defaults to `http://localhost:8080`, so it works immediately after `docker compose up`. Run the **Login** request first to populate the `{{token}}` variable for authenticated requests.

### Endpoints

| Method | Endpoint                 | Auth | Description |
|--------|--------------------------|------|-------------|
| POST   | /api/auth/register       | No   | Register with name, email, password |
| POST   | /api/auth/login          | No   | Returns JWT token |
| GET    | /api/projects            | Yes  | List projects the user owns or has tasks in |
| POST   | /api/projects            | Yes  | Create project (owner = current user) |
| GET    | /api/projects/:id        | Yes  | Get project details + its tasks |
| PATCH  | /api/projects/:id        | Yes  | Update name/description (owner only) |
| DELETE | /api/projects/:id        | Yes  | Delete project and all tasks (owner only) |
| GET    | /api/projects/:id/stats  | Yes  | Task counts by status and assignee |
| GET    | /api/projects/:id/tasks  | Yes  | List tasks — `?status=`, `?assignee=`, `?page=`, `?limit=` |
| POST   | /api/projects/:id/tasks  | Yes  | Create task (assign via `assignee_email`) |
| PATCH  | /api/tasks/:id           | Yes  | Update task (owner, creator, or assignee) |
| DELETE | /api/tasks/:id           | Yes  | Delete task (owner or creator only) |
| GET    | /api/health              | No   | Health check |

### Error Format

**Validation (400):** `{ "error": "validation_failed", "fields": { "field": "message" } }`

**Auth/Not Found/Forbidden/Server (401/403/404/500):** `{ "error": "message" }`

---

## Running Tests

### Unit Tests (no database required)

```bash
cd backend
go test ./internal/... -v
```

33 tests covering:
- **Model validation:** UUID format, date format, status/priority enums, empty string normalization
- **AuthService:** register, login, duplicate email, wrong password, user not found
- **ProjectService:** owner-only update/delete, forbidden for non-owners
- **TaskService:** update/delete authorization for owner, creator, assignee, and unrelated users

### Integration Tests (requires PostgreSQL)

```bash
cd backend
DATABASE_URL="postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable" go test ./tests/integration/ -v
```

18 tests covering all API endpoints end-to-end against a real database:
- Auth: register, login, validation errors, protected route without token
- Projects: CRUD, ownership enforcement, stats
- Tasks: CRUD, status filtering, authorization (owner/creator/forbidden)
- Validation: invalid UUIDs, invalid query params

---

## Project Structure

```
taskflow-shrinish/
├── docker-compose.yml          # PostgreSQL + API server
├── .env.example                # Environment variable template
├── postman_collection.json     # Postman v2.1 collection for all endpoints
├── README.md
└── backend/
    ├── Dockerfile              # Multi-stage: go build → alpine runtime
    ├── cmd/server/main.go      # Entry point, DI wiring, graceful shutdown
    ├── internal/
    │   ├── config/             # Environment-based configuration
    │   ├── database/           # Connection pool + seed data
    │   ├── handler/            # HTTP handlers (thin — delegates to service)
    │   ├── httputil/           # Shared response helpers (JSON, WriteError)
    │   ├── middleware/         # JWT auth, request logging, body limit
    │   ├── model/              # Structs + validation (User, Project, Task)
    │   ├── repository/         # SQL queries via pgx
    │   ├── router/             # Route definitions + middleware chain
    │   └── service/            # Business logic, authorization, AppError
    ├── migrations/             # Up/down SQL migration files
    └── tests/integration/      # Integration tests (auth, CRUD)
```

---

## What I'd Do With More Time

### Project Membership & Access Control
- **`project_members` table** — right now any authenticated user can create tasks in any project and be assigned to any task. A membership model (`project_members` with a `role` column: `owner`, `admin`, `member`) would restrict access properly.
- **`GET /api/projects/:id/members`** — endpoint to list project members, enabling a user dropdown for assignee selection instead of typing raw emails.
- **RBAC via middleware** — authorization checks currently live in the service layer (e.g., "is this user the project owner?"). A cleaner approach would be an authorization middleware or interface that resolves the user's role for a given resource before the request hits the service. This would centralize permission logic, make it testable in isolation, and keep services focused on business rules rather than access control.

### Security
- **Rate limiting** on auth endpoints to prevent brute-force attacks
- **Refresh tokens** with shorter access token expiry (15 min) for better security posture
- **Input sanitization** beyond validation for defence in depth

### Testing
- **CI/CD pipeline** with GitHub Actions (lint, test, build on every PR)

### Features
- **Real-time updates** via WebSocket/SSE so changes from other users appear instantly
- **Task comments** and activity history
- **Search** across projects and tasks

### Code Quality
- **sqlc** for type-safe SQL query generation (eliminating manual Scan calls)
- **Redis caching** for frequently accessed project/task lists
