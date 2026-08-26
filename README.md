# Yellow (Tabbycat-Go)

Yellow is a modern rebuild of the Tabbycat tournament tab software in Go, utilizing a per-tournament SQLite database architecture, a pure Go build pipeline (zero CGO toolchain requirements), and a React SPA compiled directly into a single static binary.

---

## 1. Architecture & Key Design Choices

* **Pure Go SQLite Engine:** Uses `modernc.org/sqlite` with strict single-writer discipline (`SetMaxOpenConns(1)`), `PRAGMA foreign_keys = ON;`, `PRAGMA journal_mode = WAL;`, and `PRAGMA busy_timeout = 5000;`.
* **Standard Library HTTP Routing:** Native Go 1.22+ `http.ServeMux` routing with route-level admin authorization and write blocking.
* **Single-Binary Distribution:** React frontend (Vite + TypeScript + Tailwind CSS design system) compiled to `web/dist` and embedded directly into the Go server executable via `go:embed` inside `internal/embed/dist`.
* **Tenant & Data Isolation:** Every tournament is self-describing; its schema, settings, round records, ballots, feedback, and access tokens live within its own SQLite file (`tournaments/<slug>.db`).
* **Participant Portals:** Unguessable UUID/random tokens provide secure, role-scoped participant and adjudicator dashboards.

---

## 2. Feature Matrix

### Setup & Organization
* **Hierarchy & Divisions:** Institutions, Teams, Speakers, and Adjudicators with Novice, ESL, and EFL division flags.
* **Break Categories:** Configurable break categories with custom size thresholds, base point cutoffs, and eligibility rules.
* **Custom Score Scales & Precedence:** Configurable speaker score bounds (`score_min`, `score_max`) and multi-tier ranking precedence (e.g. `["points", "speaker_points", "margin"]`).

### Matchmaking & Panel Allocation
* **Kuhn-Munkres (Hungarian) Side Balancing:** Deterministic positional history optimization across 2-sided and 4-sided debate formats.
* **Power-Pairing & Pull-Ups:** Odd-bracket pull-up/pull-down pairing with standby team admission and surplus handling.
* **Panels & Trainees:** Strength-balanced panel allocation tracking bracket importance (top debates receive highest-rated chairs and panelists) with lowest-quartile trainee distribution.
* **Conflict (Clash) Engine:** Personal, institutional, and historical conflict registration with hard and soft weight penalties.
* **Interactive Allocations Board:** Drag-and-drop team and adjudicator allocations board with real-time clash overlays.

### Balloting & Verification
* **Consensus & Split Ballots:** Support for consensus scoring and per-judge individual split ballots.
* **Double-Entry Verification:** Two independent tab-room draft entries required to confirm, with automatic discrepancy diff highlighting on conflicts.
* **Ballot Registry:** Live round ballot status dashboard tracking drafts, submissions, discrepancies, and confirmations.

### Attendance & Check-ins
* **Admin Dashboard Check-ins:** Direct check-in toggles, search filtering, attendance counters, and bulk "Check In All" / "Undo All" controls.
* **Self-Service QR Check-ins:** Mobile-friendly QR code check-in URLs (`/checkin/<token>`).
* **Round Availability Synchronization:** One-click attendance sync into round availability exclusions.

### Feedback Engine
* **Questionnaire Builder:** Dynamic question builder supporting numeric scale, text, checkbox, and select inputs with custom sequencing.
* **Bidirectional Scoping:** Scoped evaluations (team→adjudicator, panelist→chair, chair→panelist).
* **Dynamic Rating Recalculation:** Configurable test score and peer feedback weighting recalculating live judge ratings.

### Elimination Brackets
* **Break Calculation:** Automated qualifier ranking across open and novice categories with bubble detection.
* **Knockout Bracket Visualizer:** Sequential tree visualizer (Octos, Quarters, Semis, Finals) with seeding-preserved advancement progression.

### Archive & Read-Only Locks
* **Write Blocker Middleware:** Automatically protects archived tournament records from modification (`POST`, `PUT`, `DELETE`).
* **Archive Upload (`POST /api/archive/upload`):** Raw `.db` archive ingestion with automatic schema validation and connection eviction.

---

## 3. Running & Developing Yellow

### Prerequisites
* Go 1.22+
* Node.js & `pnpm` (for frontend building)

### Run the Standalone Server
```bash
# Start server with embedded frontend
go run ./cmd/server -port 8080
```
Then open: **[http://localhost:8080](http://localhost:8080)** (Admin password: `admin`)

### Run Automated Tests
```bash
# Execute the full Go test suite
go test -v ./...
```

### Frontend Development Mode
1. Start the Go backend API server:
   ```bash
   go run ./cmd/server -port 8080
   ```
2. In a separate terminal, run the Vite development server:
   ```bash
   cd web
   pnpm dev
   ```
3. Open `http://localhost:5173/` (requests to `/api` proxy automatically to port 8080).

### Build for Production
```bash
# 1. Build frontend assets
(cd web && pnpm build)

# 2. Compile static Go binary
go build -o yellow ./cmd/server
```

---

## 4. License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

