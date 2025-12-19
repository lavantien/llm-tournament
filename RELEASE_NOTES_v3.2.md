# v3.2 (2025-12-20)

This release is focused on UI correctness and compaction.

## Highlights

### Stats Chart Fix
- Ensures the Chart.js container has a stable height so stacked score bars render reliably.

### UI Compaction & Alignment
- Further reduced left navigation rail width to reclaim content space.
- Centered manual evaluation score selection and action buttons.

## Verification

- `CGO_ENABLED=1 go test ./... -v -race -cover`
- `npm run screenshots`

