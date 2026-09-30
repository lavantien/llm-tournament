# Vendored browser scripts

Served locally so the app runs fully offline. Update by re-downloading the
pinned version and recording it here.

| File             | Version | Source                                                                 |
| ---------------- | ------- | ---------------------------------------------------------------------- |
| marked.umd.js    | 18.0.14 | https://cdn.jsdelivr.net/npm/marked@18.0.14/lib/marked.umd.js          |
| chart.umd.min.js | 4.5.1   | https://cdnjs.cloudflare.com/ajax/libs/Chart.js/4.5.1/chart.umd.min.js |

Note: the previously referenced `marked@18.0.14/marked.min.js` CDN path does
not exist in the marked 18 package (files live under `lib/`), which silently
broke markdown preview before these files were vendored.
