# UI Redesign Rollout — Requirements Contract

## Goal
Apply the approved “Neon Glass Foundry” visual system across the existing SSR Go-template UI (no build step), delivering a cohesive futuristic desktop dashboard aesthetic on all primary screens.

## Acceptance Criteria (Testable)
1. All HTML templates in `templates/` (except `templates/design_preview.html`) reference the new stylesheet `"/templates/arena.css"`.
2. The shared nav partial `templates/nav.html` renders a **top bar** and a **left navigation rail** using stable class hooks:
   - Top bar contains an element with class `arena-topbar`.
   - Left rail contains an element with class `arena-sidebar`.
3. All HTML templates in `templates/` (except `templates/design_preview.html`) use the layout wrapper:
   - `<div class="arena-shell">` exists.
   - `<main class="arena-main">` exists.
4. `CGO_ENABLED=1 go test ./... -v -race -cover` passes.

## Non-goals
- Rewriting application behavior, handlers, routing, or WebSocket protocol.
- Introducing a build pipeline, SPA framework, or bundler.
- Perfect mobile responsiveness.

## Constraints
- Keep Go templates and SSR architecture.
- CDN links allowed, but the UI must remain usable with font fallbacks.
- Do not edit generated directories (e.g., `gen/`).

## Verification Plan
1. Run unit tests after each TDD step:
   - `CGO_ENABLED=1 go test ./... -v`
2. Final verification:
   - `CGO_ENABLED=1 go test ./... -v -race -cover`

