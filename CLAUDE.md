# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

Asistente Personal — microservicio Go que funciona como backend de un asistente personal accesible por WhatsApp. Parsea gastos, guarda notas con busqueda semantica, mantiene memoria de conversaciones, y ejecuta acciones proactivas via cron jobs. Se conecta con n8n como orquestador de workflows.

## Tech Stack

- **Language**: Go 1.24+
- **Framework**: Gin (via abstraccion framework-agnostic en `web/`)
- **Database**: SQLite con WAL mode o PostgreSQL (configurable)
- **AI**: Claude API (Anthropic) o OpenAI para parseo de lenguaje natural y embeddings
- **Integrations**: Google Sheets, Google Calendar, Notion, Obsidian, WhatsApp Business, GitHub, Jira, Spotify, Todoist, Gmail, ClickUp
- **Deploy**: Docker multi-stage, Oracle Cloud Free Tier
- **Orchestrator**: n8n (self-hosted) para workflows de WhatsApp y cron visual

## Development Commands

```bash
make run          # go run ./cmd
make build        # CGO_ENABLED=1 go build -o asistente ./cmd
make test         # go test -race ./...
make test-cover   # test + coverage report
make vet          # go vet ./...
make docker       # docker compose up -d --build asistente
make docker-all   # docker compose up -d --build (full stack)
make docker-down  # docker compose down
make clean        # remove binaries and coverage
```

## Architecture

Layered architecture siguiendo el patron del api-template:

```
cmd/                    ← Entry point + wiring
├── main.go             ← main() → NewApp → Run (3 lineas)
├── server.go           ← App struct, NewApp(), Run(), Close(), signal handling
├── clients.go          ← Clients struct + NewClients() + factory functions
├── controller.go       ← Controllers struct + NewControllers() con nil-checks
├── service.go          ← NewMemoryService, NewFinanceService, NewEmbedder
├── scheduler.go        ← NewScheduler() con los 4 cron jobs
└── routes.go           ← registerRoutes(), setupRoutes(), middlewareMapper()

clients/                ← External API clients (single package, prefixed types)
├── claude.go           ← ClaudeClient (Anthropic API)
├── openai.go           ← OpenAIClient (OpenAI API)
├── sheets.go           ← SheetsClient (Google Sheets)
├── calendar.go         ← CalendarClient (Google Calendar)
├── notion.go           ← NotionClient (Notion API)
├── obsidian.go         ← ObsidianVault (local filesystem)
├── whatsapp.go         ← WhatsAppClient (WhatsApp Business Cloud API)
├── github.go           ← GitHubClient (GitHub API)
├── jira.go             ← JiraClient (Jira API)
├── spotify.go          ← SpotifyClient (Spotify API)
├── todoist.go          ← TodoistClient (Todoist API)
├── gmail.go            ← GmailClient (Gmail API)
├── clickup.go          ← ClickUpClient (ClickUp API)
└── common.go           ← Shared HTTP constants

pkg/
├── domain/             ← Models, types, sentinel errors, constants, Validate() methods
├── controller/         ← HTTP handlers (decode → validate → usecase → response)
├── usecase/            ← Business logic (ProcessExpense, FallbackSearch, Compact, Scheduler)
└── service/            ← Data access interfaces + implementations (SQLite, Postgres, Sheets)
    └── sqldata/        ← SQL queries as embedded .sql files
        ├── sqlite/     ← insert/, select/, delete/ con archivos .sql individuales
        ├── postgres/   ← insert/, select/, delete/ con archivos .sql individuales
        ├── queries.go  ← Constantes SQLite via go:embed
        └── queries_pg.go ← Constantes Postgres via go:embed

internal/
├── hooks/              ← Event hook system (Register/Emit)
├── skills/             ← Skill loader (YAML frontmatter + markdown)
└── middleware/         ← Webhook auth interceptor

test/
└── mocks.go            ← MockMemoryService, MockEmbedder, MockAIProvider, MockRequest, MockClaudeServer

web/                    ← Framework-agnostic HTTP abstractions (from template)
boot/                   ← Server bootstrap (from template)
db/                     ← Migrator + SQL migrations (sqlite + postgres)
skills/                 ← Skill markdown files with YAML frontmatter
```

### Data Flow

```
HTTP Request → Controller → UseCase → Service (DB/API) → UseCase → Controller → HTTP Response
```

### Key Patterns

- **App struct**: `cmd/server.go` define `App` con `NewApp()`, `Run()`, `Close()`. `main()` son 3 lineas
- **Clients struct**: Agrupa todos los clientes externos en `cmd/clients.go`
- **Controllers struct**: Agrupa todos los controllers con nil-checks en `cmd/controller.go`
- **Framework-agnostic handlers**: Controllers usan `web.Request` / `web.Response`, no `*gin.Context`
- **Optional integrations**: Nil clients skip registration (Notion, Obsidian, Calendar, etc.)
- **Domain validation**: Request structs tienen metodos `Validate()` para datos (path traversal, fechas, URLs, max length)
- **SQL as files**: Queries en archivos `.sql` individuales organizados por operacion, embebidos via `go:embed`
- **Skills as config**: Markdown files en `skills/` inyectados como Claude system prompt
- **Context Engine**: Ingest → Assemble → Compact (multi-stage with fallback)
- **Hybrid search**: Vector + FTS5 with fallback chain
- **Graceful shutdown**: SIGINT/SIGTERM handling con deferred cleanup

## API Endpoints

| Method | Route | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/ping` | Ping/pong (built-in) |
| POST | `/api/finance/expense` | Parse expense and save to Sheets |
| GET | `/api/finance/summary` | Financial summary by period |
| POST | `/api/memory/note` | Save note with embedding |
| GET | `/api/memory/search?q=&mode=&limit=` | Search notes (fts/vector/hybrid/fallback) |
| DELETE | `/api/memory/note/:id` | Delete note |
| POST | `/api/chat` | Chat with persistent memory |
| POST | `/api/notion/page` | Create Notion page |
| GET | `/api/notion/page/:id` | Get Notion page |
| GET | `/api/obsidian/note?path=` | Read Obsidian note |
| POST | `/api/obsidian/note` | Write Obsidian note |
| GET | `/api/obsidian/notes?dir=` | List Obsidian notes |
| GET | `/api/obsidian/search?q=` | Search Obsidian vault |
| GET | `/api/calendar/today` | Today's events |
| POST | `/api/calendar/event` | Create calendar event |
| GET | `/api/github/repos` | List GitHub repos |
| GET | `/api/github/:owner/:repo/issues` | List issues |
| POST | `/api/github/:owner/:repo/issues` | Create issue |
| GET | `/api/github/:owner/:repo/pulls` | List pull requests |
| GET | `/api/jira/my-issues` | My Jira issues |
| GET | `/api/jira/issue/:key` | Get Jira issue |
| POST | `/api/jira/issue` | Create Jira issue |
| GET | `/api/spotify/playing` | Currently playing track |
| POST | `/api/spotify/play` | Resume playback |
| POST | `/api/spotify/pause` | Pause playback |
| POST | `/api/spotify/next` | Next track |
| GET | `/api/todoist/tasks` | List tasks |
| POST | `/api/todoist/task` | Create task |
| POST | `/api/todoist/task/:id/complete` | Complete task |
| GET | `/api/gmail/unread` | List unread emails |
| GET | `/api/gmail/message/:id` | Get email |
| GET | `/api/clickup/tasks` | List ClickUp tasks |
| GET | `/api/clickup/task/:id` | Get ClickUp task |
| POST | `/api/clickup/task` | Create ClickUp task |
| POST | `/api/habits/log` | Log habit |
| GET | `/api/habits/streak` | Get habit streak |
| GET | `/api/habits/today` | Today's habits |
| POST | `/api/links` | Save link |
| GET | `/api/links/search` | Search links |
| GET | `/api/projects/:name/status` | Project status summary |

## Environment Variables

See `.env.example` for the full list with defaults.

## Naming Conventions

- **Packages**: lowercase, single word (`domain`, `controller`, `usecase`, `service`)
- **Types/Structs**: PascalCase (`FinanceController`, `MemoryUseCase`, `ParsedExpense`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **File names**: snake_case (`memory_sqlite.go`, `finance_sheets.go`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **JSON fields**: snake_case (`session_id`, `paid_by`, `amount_usd`)
- **Sentinel errors**: `Err` prefix PascalCase (`ErrStoreOpen`, `ErrClaudeAPI`, `ErrValidation`)
- **Client types**: Prefijo del servicio (`ClaudeClient`, `SheetsClient`, `GitHubIssue`, `CalendarEvent`)
- **SQL constants**: PascalCase para SQLite (`SaveMemory`), prefijo `PG` para Postgres (`PGSaveMemory`)

## Error Handling

- All sentinel errors in `pkg/domain/errors.go`
- Use `domain.Wrap(sentinel, detail)` or `domain.Wrapf(sentinel, cause)` to wrap errors
- Compare errors with `errors.Is()`, never string comparison
- Controllers return typed response structs, never `map[string]any`
- Validation errors use `domain.ErrValidation` sentinel

## Validation

Request structs en `pkg/domain/` tienen metodos `Validate() error` para validaciones de datos:

- **Path traversal**: `ObsidianNoteRequest`, `ValidatePath()` — rechazan `..` y paths absolutos
- **Date format**: `CalendarEventRequest` (RFC3339), `TodoistCreateTaskRequest` (YYYY-MM-DD)
- **Cross-field**: `CalendarEventRequest` — end debe ser posterior a start
- **URL format**: `LinkSaveRequest` — solo http/https con host valido
- **Max length**: `ChatRequest`, `NoteRequest` — previenen payloads excesivos

Controllers llaman `payload.Validate()` y devuelven 400 con el error message.

## Testing

- 190+ tests across 20+ test files
- `testify` for assertions (`assert`, `require`)
- AAA pattern without comments, separated by blank lines
- Table-driven tests for multiple scenarios
- Mocks in `test/mocks.go` (MockMemoryService, MockEmbedder, MockAIProvider, MockRequest, MockClaudeServer)
- See `TESTING.md` for full guide

## Important Notes

- `web/` and `boot/` are the template base — do not modify unless necessary
- Controllers never use `*gin.Context` directly
- All integrations are optional — nil if not configured
- SQLite uses WAL mode; Postgres uses `tsvector` with spanish dictionary for FTS
- SQL queries live in `.sql` files under `pkg/service/sqldata/`, never inline strings
- Skills are hot-reloaded from disk on every chat request
- `credentials.json` (Google service account) never committed to repo
- Cron jobs run in isolated goroutines
