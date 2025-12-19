# v3.1 (2025-12-19)

This is a small polish release after `v3.0`, focused on tightening the Arena UI layout and keeping the README UI Tour screenshots in sync.

## Highlights

### Arena UI Compaction
- Thinner left navigation rail for more horizontal space.
- Sticky headers/footers and title/tool rows now use compact flex layouts (buttons stay on the same row when there’s room).
- Dropdowns (`<select>`) and file inputs are styled to match the Arena theme.
- “Scroll to top / bottom” buttons use the sidebar space on shell pages (and stay bottom-right on solo pages).

### Docs: Auto-generated UI Screenshots
- `npm run screenshots` regenerates `assets/ui-*.png` using a deterministic demo server + Playwright.

## Verification

- `CGO_ENABLED=1 go test ./... -v -race -cover`
- `npm run screenshots`

