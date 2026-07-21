# GoTabs (Tabbycat-Go)

GoTabs is a tournament tab software rebuild in Go, utilizing a per-tournament SQLite database architecture, a pure Go build pipeline (no CGO dependencies), and a modern React client SPA compiled directly into the executable binary.

---

## 1. Project Overview & Architecture

* **Database Engine:** Pure Go SQLite driver (`modernc.org/sqlite`). Connection pooling enforces a single-writer constraint (`SetMaxOpenConns(1)`) to eliminate `SQLITE_BUSY` contention.
* **Backend Router:** Standard Go 1.22+ `http.ServeMux` routing pattern.
* **Frontend Shell:** React (Vite + TS + Tailwind CSS) client. Assets compile to `web/dist` and are embedded directly into the Go executable via `go:embed` inside `internal/embed/dist`.
* **Tenant Isolation:** Every tournament's settings, schemas, records, and access tokens live within its own SQLite file (`tournaments/<slug>.db`).
* **Visual Theme:** Light-monochrome theme with zinc-colored backdrops and cards, responsive layouts, and modal workflows.

---

## 2. What Has Been Completed

### Phase 0: Scaffolding & Dynamic DB Manager
* Platform database `tournaments/global.db` tracks tournament registries and archives.
* Connection manager [internal/db/manager.go](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/db/manager.go) loads SQLite files dynamically, enforcing read-only modes if flagged.

### Phase 1: Tournament Operations & REST APIs
* REST handlers for Institutions, Teams, Speakers, Adjudicators, and Rounds in [internal/api/tournament.go](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/api/tournament.go).
* Bulk CSV upload parsers in [internal/api/csv.go](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/api/csv.go).
* Kuhn-Munkres (Hungarian algorithm) side position balance solver and power-pairing matchmaking draws in [internal/draw/](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/draw/).
* Token-resolved participant and adjudicator route schedules in [internal/api/token.go](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/api/token.go) and [web/src/App.tsx](file:///Users/dev-isolated/Desktop/Projects/GoTabs/web/src/App.tsx).

### Phase 2: Archive & Read-Only Locks
* Upload parser for ingested database archives in [internal/api/archive.go](file:///Users/dev-isolated/Desktop/Projects/GoTabs/internal/api/archive.go).
* Write blocker middleware intercepting mutating queries (`POST`, `PUT`, `DELETE`) on archived paths returning `403 Forbidden`.

### Final UI Adjustments (Round Controls & Standings)
* Added round configuration toggles (Draw Released, Silent Round, Results Released) synced via `PUT /api/t/{slug}/rounds/{round_id}`.
* Added participant standings card displaying ranking and scores on `/p/:token`.
* Null-safety JSON serialization guards returning `[]` instead of `nil` for empty states.

---

## 3. What Needs to Be Done (Next Steps)

### 1. Organizer Endpoint Authorization Middleware (High Priority)
* Currently, tournament admin REST endpoints (e.g., `/api/t/{slug}/institutions` POST/DELETE) do not validate whether the organizer is logged in. They rely purely on frontend password gating.
* **Task:** Implement an auth middleware in Go that inspects the `admin_session` cookie and wraps all tournament admin/writer routes.

### 2. Phase 3: Cloud Replication & OAuth2 Topology
* **LiteFS/Litestream Sync:** Integrate backup/sync replication to ship SQLite tournament databases off to cloud buckets automatically.
* **OAuth2 Authentication:** Set up Auth0, Google, or GitHub social logins for organizers instead of the hardcoded `"admin"` credentials.
* **Topology Segregation:** Allow splitting the platform registry from the running tournaments to support distributed container instances.

---

## 4. Run & Development Instructions

### Run the Compiled Static Application
```bash
./server -port 8080
```
Then visit: `http://localhost:8080/`

### Frontend Live Development Mode
1. Run the Go server:
   ```bash
   ./server -port 8080
   ```
2. Navigate to the frontend directory:
   ```bash
   cd web
   ```
3. Start the Vite dev server:
   ```bash
   pnpm run dev
   ```
4. Access the hot-reloading client at `http://localhost:5173/`. All API calls are automatically proxied to the Go server on port 8080.
