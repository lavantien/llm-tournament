# LLM Tournament Arena

[![Coverage](./coverage-badge.svg)](./coverage.html)
[![CI](https://github.com/lavantien/llm-tournament/workflows/CI/badge.svg)](https://github.com/lavantien/llm-tournament/actions)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat&logo=sqlite&logoColor=white)](https://sqlite.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A local-first benchmarking arena for evaluating and comparing Large Language Models (LLMs) with manual scoring.

**Highlights**

- SQLite-backed, single-binary Go server with SSR templates + WebSockets (`:8080`)
- Prompt suites, profiles, models, results grid, and analytics
- Fully offline: no third-party APIs, no API keys

**UI Stack**

- Tailwind CSS v4.3.3 + DaisyUI v5.7.47 (0% custom CSS)
- Built-in DaisyUI components and themes (coffee)
- Industry-standard utility-first styling approach
- Zero maintenance custom CSS codebase

## Table of Contents

- [1. Quick Start](#1-quick-start)
- [2. UI Design](#2-ui-design)
- [3. Features](#3-features)
- [4. Architecture](#4-architecture)
- [5. Tech Stack](#5-tech-stack)
- [6. Installation](#6-installation)
- [7. Usage Tutorial](#7-usage-tutorial)
- [8. Development](#8-development)
- [9. Testing](#9-testing)
- [10. Troubleshooting](#10-troubleshooting)
- [11. API Reference](#11-api-reference)
- [12. Project Structure](#12-project-structure)
- [13. Environment Variables](#13-environment-variables)
- [14. Documentation Guidelines](#14-documentation-guidelines)
- [15. License](#15-license)
- [16. Contact](#16-contact)

## 1. Quick Start

```bash
git clone https://github.com/lavantien/llm-tournament.git
cd llm-tournament
make run
```

Open http://localhost:8080 (data is stored in `data/tournament.db` by default).

No `make`? Run directly:

```bash
CGO_ENABLED=1 go run .
```

PowerShell:

```powershell
$env:CGO_ENABLED=1; go run .
```

[↑ Back to top](#table-of-contents)

## 2. UI Design

**Updated: Tailwind v4 + DaisyUI v5 (Zero Custom CSS)**

The UI has been migrated to use **100% pure Tailwind v4 + DaisyUI v5** components. See [DESIGN_CONCEPT.md](DESIGN_CONCEPT.md) for complete design specifications and [DESIGN_ROLLOUT.md](DESIGN_ROLLOUT.md) for detailed migration plan.

**Key Design Decisions:**

- Zero custom CSS - all styling uses Tailwind utilities or DaisyUI semantic components
- Built-in DaisyUI `cyberpunk` theme provides dark backgrounds with neon accents
- Tailwind v4 built-in animations (`animate-spin`, `animate-ping`, `animate-pulse`) replace custom keyframes
- Dynamic score theming uses Tailwind arbitrary values (`bg-[#color]`) instead of CSS variables
- Glass panels use DaisyUI `.card` components without custom glow effects
- Industry-standard approach using well-maintained tools (Tailwind + DaisyUI)

**Trade-offs:**

- No glass glow overlay effects (cleaner appearance)
- No grid overlay texture (cleaner background)
- Simpler animations (less dramatic, more performant)
- Standard DaisyUI components instead of custom semantic classes

[↑ Back to top](#table-of-contents)

## 3. Features

### 3.1 Manual Evaluation

- Real-time scoring on 0-100 scale (increments: 0, 20, 40, 60, 80, 100)
- Automatic model ranking with live leaderboard updates
- WebSocket-based instant updates across all clients
- State backup and rollback support
- Drag-and-drop prompt reordering and bulk operations
- Save and edit each model's response per prompt

### 3.2 Suite Management

- Independent prompt suites with isolated profiles, prompts, and results
- JSON import/export for suites and evaluation results
- Duplicate cleanup and SQLite migration support
- One-click suite switching

### 3.3 Analytics

- 12-tier classification system: Transcendental (>=3780) to Primordial (<300)
- Interactive visualizations using Chart.js
- Score distributions and tier-based model grouping
- Performance comparisons across models and prompt types

### 3.4 Interface

- Markdown editor with live preview
- Advanced search and filtering
- Copy-to-clipboard functionality
- Connection status monitoring with automatic reconnection

[↑ Back to top](#table-of-contents)

## 4. Architecture

```
Go Server (:8080)
├── HTTP Handlers
├── WebSocket Hub
└── SQLite DB
```

**High-Level System Context**

```mermaid
graph LR
    subgraph Client Side
        Browser["User Browser\n(HTML/JS/Templates)"]
    end

    subgraph Server Side
        subgraph Go Monolith
            HTTP_WS["Go HTTP & WebSocket Server"]
        end

        SQLite[("SQLite Database\n(Single Source of Truth)")]
    end

    Browser -- "HTTP Requests / WebSocket" --> HTTP_WS
    HTTP_WS -- "Reads/Writes Data" --> SQLite
```

**Layered Architecture Flow**

```mermaid
graph TD
    subgraph "Go Monolith Layers"
        Surface["1. Surface Layer\n(Templates: *.html, *.js)"]
        Handlers["2. HTTP Handlers\n(handlers/*.go)"]
        Middleware["3. Middleware Layer\n(DB, State, Render, WS)\n(middleware/*.go)"]
    end

    DB[("SQLite Database")]

    %% Main Flow based on text description
    Surface --> Handlers
    Handlers --> Middleware

    %% Data Access
    Middleware <-->|"Read/Write Schema"| DB
```

**Sequence: Manual Evaluation Flow**

```mermaid
sequenceDiagram
    participant User
    participant Browser
    participant GoServer
    participant SQLite
    participant WebSocket

    User->>Browser: Clicks score button
    Browser->>GoServer: POST /results/update
    GoServer->>SQLite: UPDATE scores SET score = ?
    SQLite-->>GoServer: Success
    GoServer->>WebSocket: Broadcast score_update
    WebSocket-->>Browser: Real-time update
    Browser->>User: Live leaderboard refresh
```

**Sequence: Prompt Management Flow**

```mermaid
sequenceDiagram
    participant User
    participant Browser
    participant GoServer
    participant SQLite

    User->>Browser: Drag prompt to reorder
    Browser->>GoServer: WS message: reorder_prompts
    GoServer->>SQLite: UPDATE prompts SET order = ?
    SQLite-->>GoServer: Success
    GoServer->>Browser: WS broadcast: update_prompts_order
    Browser->>User: Reorder animation completes
```

Request Flow: User -> Handlers -> Middleware -> SQLite -> WebSocket Broadcast

### 4.1 Bird's-Eye View

- This is a Go monolith (HTTP + WebSocket) with SQLite as a single source of truth.
- The repo is organized by "layer": surface (templates) -> HTTP handlers -> middleware (DB/state/render/ws).
- The fastest "index" is to URL handler map in `main.go:60`, and the DB schema is centralized in `middleware/database.go:58`.
- **UI Migration**: All styling now uses Tailwind v4 + DaisyUI v5 components with zero custom CSS. See [DESIGN_ROLLOUT.md](DESIGN_ROLLOUT.md) for complete migration details.

### 4.2 Where To Look In 5 Seconds

- **HTTP routes / feature entrypoint:** `main.go:60` (every user-visible feature starts as a path here).
- **HTML/JS for a page:** `templates/*.html` and `templates/*.js` (e.g. `templates/results.html`, `templates/prompt_list.html`).
- **DB tables & relationships:** `middleware/database.go:58` (schema includes `suites`, `profiles`, `prompts`, `models`, `scores`, `model_responses`).
- **Per-feature server logic:** `handlers/*.go` (files are feature-named: prompts/models/profiles/results/stats/suites).
- **WebSocket messages:** `middleware/socket.go:33` (server-side `/ws`, broadcasting and client tracking).
- **Saved model responses:** `handlers/model_response.go` (stored in `model_responses`, edited from the Evaluate page).
- **UI Components:** Tailwind v4 + DaisyUI v5. See [DESIGN_CONCEPT.md](DESIGN_CONCEPT.md) for complete component mapping.
- **Test-as-documentation:** `handlers/*_test.go`, `middleware/*_test.go`, `integration/prompts_integration_test.go`.

### 4.3 Common Feature Map

- **Prompt suites:** `main.go:76` `handlers/suites.go` (+ UI in `templates/*prompt_suite*.html`)
- **Prompts CRUD/order:** `main.go:62`/`main.go:66`/`main.go:73` `handlers/prompt.go:1` (+ reorder over WS in `middleware/socket.go:71`)
- **Models CRUD:** `main.go:63` `handlers/models.go`
- **Manual scoring/results UI:** `main.go:80`/`main.go:81` `handlers/results.go` (+ `templates/results.html`)
- **Stats/analytics:** `main.go:93` `handlers/stats.go` (+ `templates/stats.html`)

### 4.4 Search Cheats (copy/paste)

- Find a feature by URL: `rg -n '"/results"|"/stats"|"/prompts"' main.go`
- Find which handler renders a template: `rg -n "results\\.html|prompt_list\\.html" handlers`
- Find everything touching a table: `rg -n "evaluation_jobs|evaluation_history|model_responses" -S .`
- Find a websocket message type: `rg -n "update_prompts_order|results" middleware/templates -S`

[↑ Back to top](#table-of-contents)

## 5. Tech Stack

Backend: Go 1.27+, Gorilla WebSocket, Blackfriday, Bluemonday, SQLite

Frontend: HTML5, Tailwind CSS v4.3.3, DaisyUI v5.7.47, JavaScript ES6+, Chart.js 4.x, Marked.js

Security: XSS sanitization, CORS protection, input validation

[↑ Back to top](#table-of-contents)

## 6. Installation

### 6.1 Prerequisites

- Go 1.27+
- A C toolchain for CGO/SQLite (e.g., gcc/clang; on Windows install MinGW-w64/MSYS2)
- Git
- Make (optional, for convenience targets)
- Node.js and npm (for UI screenshots only)

### 6.2 Manual Evaluation (Go-only)

```bash
# Run from source
make run

# Or without make
CGO_ENABLED=1 go run .
```

Build a binary:

```bash
make build
```

Run it:

- Linux/macOS: `./release/llm-tournament`
- Windows (PowerShell): `.\release\llm-tournament.exe`

One-time migration (only if upgrading old result formats):

```bash
CGO_ENABLED=1 go run . --migrate-results
```

### 6.3 UI Installation (DaisyUI + Tailwind v4)

The UI now uses Tailwind CSS v4 + DaisyUI v5 with zero custom CSS. See [DESIGN_CONCEPT.md](DESIGN_CONCEPT.md) and [DESIGN_ROLLOUT.md](DESIGN_ROLLOUT.md) for complete migration details.

**Install dependencies:**

```bash
npm install
```

This installs:

- Tailwind CSS v4.3.3
- DaisyUI v5.7.47
- PostCSS and build tools

**Build CSS:**

```bash
npm run build:css
```

This generates `templates/output.css` from `templates/input.css` using PostCSS.

[↑ Back to top](#table-of-contents)

## 7. Usage Tutorial

This tutorial will guide you through the essential workflows of LLM Tournament Arena.

### 7.1 Your First Run

After starting the server, open `http://localhost:8080`. You'll see:

1. **Top navigation bar** - Contains links to Results, Stats, Prompts, Profiles, and Evaluate
2. **Suite selector** - On the right side of the top bar, with New/Edit/Delete buttons
3. **Prompts page** - Your starting point for managing test prompts (a default suite is created automatically)

![Prompts List](assets/ui-prompts.png)

[↑ Back to top](#table-of-contents)

### 7.2 Task: Add Your First Model

Models are added directly from the Results page:

1. Click **Results** in the top navigation bar
2. At the top of the page, you'll see a form with "Enter new model name"
3. Type the model name (e.g., "claude-3-5-sonnet-20241022", "gpt-4o", etc.)
4. Click **Add**

The model will appear in the results grid. Repeat for each model you want to evaluate.

### 7.3 Task: Create a Profile

A **Profile** is a group of models that you want to evaluate together.

1. Click **Profiles** in the top bar
2. Click **Add Profile** button
3. Enter a name for your profile
4. Select the models you want to include
5. Click **Save**

![Profiles](assets/ui-profiles.png)

### 7.4 Task: Create a Test Prompt

1. Navigate to **Prompts** in the top bar
2. Click **Add Prompt** button
3. Enter prompt details:
   - **Title**: Short descriptive name
   - **Category**: e.g., "coding", "creative-writing", "reasoning"
   - **Content**: Your test prompt (Markdown supported)
   - **Expected Answer**: Reference answer for manual comparison
4. Click **Save**

![Edit Prompt](assets/ui-edit-prompt.png)

### 7.5 Task: Run Manual Evaluation

The **Evaluate** page lets you score models one prompt at a time.

**How to access:** You typically navigate here by clicking a score cell in the Results page (see section 7.6), which automatically takes you to the evaluate page for that model and prompt.

Once on the Evaluate page, you'll see:
   - The current **model name** at the top
   - The **prompt number** (e.g., "Prompt 3 of 10")
   - The **prompt text** and **expected solution**
   - The **model's response** (which you can save)
   - **Score buttons** (0, 20, 40, 60, 80, 100)

**To score:**
   - Click a score button to select it
   - Click ✅ to submit and move to the next prompt
   - Use ⬅️➡️ buttons to navigate between prompts without scoring
   - Click ❌ to return to the **Results** page

![Evaluate](assets/ui-evaluate.png)

### 7.6 Task: View and Edit Results

The **Results** page shows your scoring grid and lets you edit individual scores.

1. **Results Grid** overview:
   - Rows show each model
   - Columns show each prompt
   - Cells show scores with color coding (green=high, red=low)
   - Total scores and progress bars on the right

2. **Edit a score**:
   - Click any score cell to go to the Evaluate page for that model×prompt combination
   - Update your score and click ✅
   - Click ❌ to return to Results

3. **Stats** (top bar link):
   - View score distributions
   - See model tier rankings
   - Compare performance across categories

![Results](assets/ui-results.png)
![Stats](assets/ui-stats.png)

### 7.7 Task: Import/Export and Suite Management

**Suite Management:**
- The **Suite selector** is in the top-right of the top navigation bar
- Use the dropdown to switch between suites
- Click **New** to create a new suite
- Click **Edit** to modify the current suite name
- Click **Delete** to remove the current suite

**Import/Export:**
- **Results page**: Contains import/export buttons for evaluation results
- **Prompts page**: Contains import/export buttons for prompt data
- Export formats use JSON for backup and portability

### 7.8 Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Navigate between cells (Results grid) | Arrow keys (when cell is focused) |
| Submit score (Evaluate page) | Enter (when score selected) |

**Note:** Other navigation elements use UI buttons (⬅️➡️ for prompts, ↑↓ for scroll to top/bottom).

### 7.9 Tips for Efficient Usage

- **Batch Operations**: Use checkboxes to select multiple prompts for bulk actions
- **Drag to Reorder**: Reorder prompts by dragging them in the list
- **Real-time Updates**: Open multiple browser tabs - they sync automatically
- **State Backup**: Save your evaluation state before long sessions
- **Suite Isolation**: Use separate suites for different evaluation projects

[↑ Back to top](#table-of-contents)

## 8. Development

This project is part of a larger development environment. For a complete setup including:

- **Shell configuration** (zsh/fish/bash with aliases, functions)
- **Go development tools** (gopls, golangci-lint, delve debugger)
- **Python environment** (pyenv, poetry, pipx)
- **Node.js tools** (nvm, npm global packages)
- **AI/LLM CLI tools** (claude-cli, openai-cli, aider)
- **Git workflows** (hooks, templates, aliases)
- **Editor configs** (Neovim/Vim/VSCode settings)

See: **[lavantien/dotfiles](https://github.com/lavantien/dotfiles)**

### 8.1 Local Development Setup

#### Prerequisites
- Go 1.27+
- Node.js 24+
- CGO-enabled toolchain (gcc/clang/MinGW)

#### Running the Development Server

```bash
# Clone the repository
git clone https://github.com/lavantien/llm-tournament.git
cd llm-tournament

# Install UI dependencies
npm install

# Build CSS (watch mode for development)
npm run build:css:watch

# Run the Go server
CGO_ENABLED=1 go run .
```

The server will start on `http://localhost:8080`.

### 8.2 Running Tests

```bash
# Full test suite with TDD guard
make test

# Run specific test package
CGO_ENABLED=1 go test ./handlers -v -race -cover

# Generate coverage report
CGO_ENABLED=1 go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 8.3 Building CSS

```bash
# One-time build
npm run build:css

# Watch mode (rebuilds on changes)
npm run build:css:watch
```

### 8.4 Generating Screenshots

```bash
npm install
npm run screenshots:install
npm run screenshots
```

Screenshots are saved to `assets/ui-*.png`.

CI does not generate screenshots (the Playwright job was removed for speed). Regenerate them locally with the commands above before pushing UI changes.

### 8.5 Additional Documentation

- UI design and migration: [DESIGN_CONCEPT.md](DESIGN_CONCEPT.md), [DESIGN_ROLLOUT.md](DESIGN_ROLLOUT.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

[↑ Back to top](#table-of-contents)

## 9. Testing

```bash
# Run all tests with verbose output, race detection, and coverage
make test

# Quieter run that also writes coverage.out
make testbrief

# Lint + testbrief + per-function coverage report
make check

# Manual test run
CGO_ENABLED=1 go test ./... -v -race -cover
```

### 9.1 Testing Methodology (UI Components)

**Important**: Tailwind/DaisyUI classes are strings in your templates—no runtime JS needed for unit tests. You verify classes are present in rendered HTML without applying actual CSS.

**Go SSR Testing**: Full SSR flow verification with `httptest`. Test handlers execute templates with data and verify DaisyUI classes are present in the rendered output.

**httptest Integration**: HTTP handler testing with rendered HTML output. Verify template rendering with real data structures.

**Visual Regression**: Use existing screenshot system to compare before/after UI states.

### 9.2 Coverage

Package-level statement coverage from `CGO_ENABLED=1 go test ./... -coverprofile coverage.out`:

| Package | Coverage |
| --- | ---: |
| llm-tournament | 100.0% |
| llm-tournament/handlers | 100.0% |
| llm-tournament/integration | - |
| llm-tournament/middleware | 100.0% |
| llm-tournament/templates | 100.0% |
| llm-tournament/testutil | 100.0% |
| llm-tournament/tools/screenshots/cmd/demo-server | 100.0% |
| **Total** | **100.0%** |

[↑ Back to top](#table-of-contents)

## 10. Troubleshooting

- `CGO_ENABLED=1` set but build fails: install a working C compiler toolchain (CGO required it for SQLite).
- Port already in use: stop conflicting process (Go server currently listens on `:8080` in `main.go`).
- DB issues: default DB is `data/tournament.db`; you can point to another file with `--db <path>`.
- **DaisyUI classes not rendering**: Verify `tailwind.config.js` includes DaisyUI plugin and `npm run build:css` has been run.

[↑ Back to top](#table-of-contents)

## 11. API Reference

### 11.1 Core Endpoints

- GET /prompts - Prompts list (default route)
- GET /results - Results and scoring
- GET /profiles - Profile management
- GET /stats - Analytics dashboard
- GET /evaluate?model={name}&prompt={index} - Manual scoring page
- POST /save_model_response - Save a model's response text for a prompt
- WS /ws - WebSocket connection

[↑ Back to top](#table-of-contents)

## 12. Project Structure

```
llm-tournament/
├── main.go              # Entry point, routing, server setup
├── handlers/            # HTTP handlers (models, prompts, results, stats, suites, profiles)
├── middleware/          # Business logic (database, WebSocket, state, rendering)
├── templates/           # HTML, CSS, JavaScript
├── assets/              # UI screenshots and static images
├── data/                # SQLite database
├── tailwind.config.js    # Tailwind v4 + DaisyUI v5 configuration
└── postcss.config.js    # PostCSS configuration
```

**UI-Specific:**

- `templates/input.css` - Tailwind + DaisyUI imports only (zero custom CSS)
- `templates/output.css` - Generated CSS file (PostCSS output)
- `templates/*.html` - All HTML templates using DaisyUI components

[↑ Back to top](#table-of-contents)

## 13. Environment Variables

- CGO_ENABLED=1 (required for SQLite)

[↑ Back to top](#table-of-contents)

## 14. Documentation Guidelines

When editing documentation files, be aware that several files are automatically validated by tests and CI scripts. See [DOCUMENTATION_ENFORCEMENT.md](DOCUMENTATION_ENFORCEMENT.md) for:

- List of enforced documentation files (README.md, DESIGN_CONCEPT.md, design_preview.html)
- Required sections and formats for each file
- How to update documentation without breaking automation
- Troubleshooting common mistakes

**Quick reference:**

- `README.md` - Coverage table enforced by `scripts/update_coverage_table.py`
- `DESIGN_CONCEPT.md` - Section headers enforced by `design_preview_test.go`
- `design_preview.html` - Required elements enforced by `design_preview_test.go`

**Pre-commit hook (recommended):**

```bash
# Install automatic documentation verification before commits
cp scripts/pre-commit .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit
```

This will automatically run `make verify-docs` when you commit documentation changes.

**To verify documentation changes:**

```bash
# Run specific enforcement test
CGO_ENABLED=1 go test -run TestDesignConceptAndPreview_ExistAndStructured -v

# Update coverage table after editing README
make update-coverage-table

# Run full test suite
make test
```

[↑ Back to top](#table-of-contents)

## 15. License

MIT License - See [LICENSE](LICENSE) for details

[↑ Back to top](#table-of-contents)

## 16. Contact

cariyaputta@gmail.com

[↑ Back to top](#table-of-contents)
