# LLM Tournament Arena — Design Concept

## Goal
Transform the “LLM Tournament Arena” UI into a **futuristic, sleek, high-density desktop dashboard** that feels like a *sci‑fi industrial control center* (glass + neon + precise data surfaces) while staying **SSR Go templates + zero build step**.

## Constraints (Non-Negotiable)
- No SPA conversion (keep `html/template` + server-side rendering).
- No build tooling (no Vite/Webpack/Node build pipeline).
- CDN libraries only (optional; the UI must still look good with fallbacks).
- Desktop-first: optimize for large screens + information density.

## Visual Direction
**“Neon Glass Foundry”**
- Deep-space base layers with subtle grid/scanline texture.
- Glassmorphism panels with sharp industrial borders.
- Neon accents used sparingly for state, emphasis, and interactive affordances.
- Monospace “instrumentation” for numbers, IDs, tiers, and logs.

## Color Palette
**Base**
- `--bg-0`: `#070A12` (void)
- `--bg-1`: `#0B1020` (deep space)
- `--surface-0`: `rgba(16, 24, 38, 0.72)` (glass panel)
- `--surface-1`: `rgba(10, 15, 26, 0.70)` (glass inset)
- `--stroke`: `rgba(255, 255, 255, 0.14)` (hairline border)
- `--stroke-strong`: `rgba(255, 255, 255, 0.22)` (active/hover border)

**Text**
- `--text`: `rgba(244, 248, 255, 0.92)`
- `--muted`: `rgba(244, 248, 255, 0.64)`
- `--faint`: `rgba(244, 248, 255, 0.46)`

**Neon Accents**
- `--cyan`: `#19F7FF` (primary accent / focus)
- `--magenta`: `#FF2BD6` (secondary accent)
- `--violet`: `#A77BFF` (tertiary accent)
- `--green`: `#7CFF6B` (success / pass)
- `--amber`: `#FFC857` (warning / in-progress)
- `--red`: `#FF4D4D` (danger / failed)

**Data / Charts**
- Use the existing score semantics (0/20/40/60/80/100), but render them as:
  - “Heat chips” (bordered pills) + compact stacked bars.
  - Chart.js theme: dark gridlines, glowing dataset colors, bold tooltip contrast.

## Typography
- UI: **Inter** (clean, modern, highly legible at dense sizes)
- Data/Code: **JetBrains Mono** (scores, tiers, IDs, logs, timestamps)
- Fallbacks: `system-ui, Segoe UI, Roboto, Arial, sans-serif` and `ui-monospace, Consolas, Menlo, monospace`

## UI Elements
**Layout**
- **Top Bar**: product identity + suite context + live connection state.
- **Left Rail**: dense navigation (Results / Stats / Prompts / Profiles / Evaluate / Settings), quick actions, “hotkeys” hints.
- **Main Grid**: 2-column dashboard (table-heavy left; charts/insights right).

**Panels / Cards (Glassmorphism + Industrial)**
- `glass-panel`: blurred, gradient glass with border + subtle inner glow.
- `glass-inset`: recessed areas for tables, logs, and editors.
- Use “corner accents” (small pseudo-elements) for a cyber-instrument feel.

**Buttons**
- `neon-button`: gradient border, soft glow on hover, crisp focus ring.
- Variants:
  - Primary (cyan glow): Run / Save / Evaluate
  - Ghost (low emphasis): Cancel / Back
  - Danger (red): Delete / Reset

**Tables**
- Sticky header, tight row height, monospace numeric columns.
- Row hover highlights with faint neon edge.
- Status chips: Tier, Pass/Fail, Job state.

**Consensus Graph (Concept)**
- A compact “Judge Matrix” card:
  - 3 judge rows (Claude / GPT / Gemini) with confidence bar + verdict chip.
  - A “Consensus” gauge (weighted median) with a neon sweep.
  - Hover reveals judge reasoning (accordion/log panel).

**System Feedback**
- Connection pill (WS): Connected / Reconnecting / Failed.
- Toasts for save / copy / job started.
- “Telemetry” footer: last broadcast time, active suite, job queue depth.

## Libraries (CDN Only)
- Fonts: Google Fonts (Inter + JetBrains Mono).
- Icons (optional): Font Awesome 6 (or inline SVGs for zero dependency).
- Charts: Chart.js (already used in `templates/stats.html`).
- Animation (optional): Anime.js (only if needed; CSS animations preferred).

## Implementation Notes (Next Phase, After Approval)
- Introduce a new theme layer (CSS variables + shared classes) and progressively apply it to:
  - `templates/nav.html` (top bar + density-focused layout)
  - `templates/results.html`, `templates/prompt_list.html`, `templates/profiles.html`, `templates/stats.html`
- Keep markup changes minimal; focus on:
  - consistent “panel” wrappers
  - normalized buttons/inputs
  - table styling and spacing

