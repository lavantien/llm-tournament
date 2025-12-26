# UI Redesign Rollout — Tailwind v4 + DaisyUI v5 (Zero Custom CSS)

## Goal

Migrate from 1,065 lines of custom CSS to **100% pure Tailwind v4 + DaisyUI v5** built-in solutions. Deliver a clean, maintainable UI using only utility classes and DaisyUI semantic components while maintaining SSR Go template architecture.

## Acceptance Criteria (Testable)

1. **Zero custom CSS** in `templates/input.css` - only imports for Tailwind and DaisyUI
2. All HTML templates in `templates/` reference DaisyUI v5 component classes and Tailwind v4 utility classes
3. Tailwind v4 + DaisyUI v5 configuration properly set up in `tailwind.config.js` and `postcss.config.js`
4. `CGO_ENABLED=1 go test ./... -v -race -cover` passes after each phase
5. Full SSR flow verified with `httptest` against Go handlers rendering templates
6. Visual regression testing with existing screenshot system
7. No custom CSS variables, keyframes, or layer directives beyond Tailwind/DaisyUI imports

## Non-goals

- No build pipeline changes (keep server-side rendering)
- No SPA conversion (maintain Go template architecture)
- No framework additions (Tailwind + DaisyUI only)
- No visual effect preservation requiring custom CSS (accept redesign trade-offs)

## Constraints

- **Zero custom CSS rule**: All styling must use Tailwind v4 utilities or DaisyUI v5 components
- Built-in DaisyUI themes only (no custom theme definitions)
- Tailwind built-in animations only (no custom `@keyframes`)
- No CSS pseudo-elements (`::before`, `::after`) for visual effects
- No complex multi-layer gradients in CSS (use DaisyUI theme colors or Tailwind arbitrary values)
- No custom CSS variables (use Tailwind arbitrary values or DaisyUI color semantics)
- Desktop-first optimization maintained (from original design concept)
- SSR architecture preserved (`html/template` package)
- Testing verification through Go template rendering with `httptest` (not CSS unit tests)

## Trade-offs Required (Visual Redesign)

To achieve 0% custom CSS, we accept these design changes:

| Original Feature | Built-in Solution | Visual Impact |
|-----------------|-------------------|---------------|
| **Glass panel glow effects** (`::before` pseudo-elements with radial gradients) | DaisyUI `.card` with `shadow-lg` | ✗ No glow overlay, cleaner appearance |
| **3-layer body background gradients** (radial + linear + grid pattern) | DaisyUI `cyberpunk` theme | ✗ Different gradient pattern, similar dark aesthetic |
| **Grid overlay texture** (CSS pattern with mix-blend-mode) | Removed entirely | ✗ No grid texture, clean background |
| **Custom animations** (`slowGlow`, `shimmer`, `pulse-connected`) | Tailwind `animate-pulse`, `animate-ping`, `animate-spin` | ✗ Simpler animations, less dramatic |
| **Dynamic score colors** (CSS variables `--score-color-X`) | Tailwind arbitrary values `bg-[#color]` or DaisyUI semantic colors | ✗ More verbose HTML but flexible |
| **Complex gradient borders** (cyan→magenta→violet) | DaisyUI `.btn-primary`, `.btn-info`, `.btn-accent` | ✗ Single color per button variant |
| **Custom button hover effects** (transform translateY, filter saturate) | DaisyUI `.btn` with built-in hover states | ✗ Simpler hover transitions |

## Verification Plan

### Foundation Phase
1. Install DaisyUI v5: `npm install daisyui@latest --save-dev`
2. Replace `tailwind.config.js` with minimal config using DaisyUI plugin
3. Verify `postcss.config.js` compatibility with v5
4. Delete all custom CSS from `templates/input.css` (keep only imports)
5. Run `npm run build:css` and verify output generation
6. Run `CGO_ENABLED=1 go test ./... -v` to ensure no regressions

### Component Migration Phase (All at Once)
1. Create mapping table of all custom classes to DaisyUI/Tailwind equivalents
2. Update all template files to use DaisyUI components
3. Verify semantic HTML structure (no custom class names)
4. Run full test suite with coverage: `CGO_ENABLED=1 go test ./... -v -race -cover`

### Page-by-Page Verification Phase (Easiest → Hardest)

Run `httptest` integration tests for each page after migration:

**Tier 1 - Simplest:**
1. `nav.html` - Verify `.btn-primary`, `.menu` rendering
2. `delete_prompt.html`, `delete_prompt_suite.html`, `delete_model.html`, `delete_profile.html` - Verify `.card`, `.input`, `.btn`
3. `bulk_delete_prompts.html`, `confirm_refresh_results.html`, `import_prompts.html`, `import_results.html`, `move_prompt.html` - Verify form components

**Tier 2 - Medium:**
4. `new_prompt_suite.html` - Verify `.card`, `.input`, `.select`, `.btn`
5. `edit_prompt_suite.html`, `edit_prompt.html`, `edit_profile.html`, `edit_model.html` - Verify forms with textarea
6. `reset_prompts.html`, `reset_profiles.html`, `reset_results.html` - Verify `.btn btn-error`
7. `import_error.html` - Verify `.alert` components

**Tier 3 - Higher:**
8. `profiles.html` - Verify `.card`, `.table`, `.btn btn-square`, `.badge`, `.progress`
9. `prompt_list.html` - Verify drag-drop UI, tables, filtering
10. `results.html` - Verify large tables, badges, `.table-zebra`
11. `settings.html` - Verify `.input`, `.select`, `.checkbox`, `.range`

**Tier 4 - Hardest:**
12. `stats.html` - Verify chart rendering with DaisyUI colors
13. `evaluate.html` - Verify complex score theming, evaluation form
14. `design_preview.html` - Verify multiple `.card` components, layout

### Testing Methodology

**Go SSR Testing:**
```go
// Test template rendering with real data
func TestPageTemplatesWithDaisyUI(t *testing.T) {
    // Render each template with test data
    // Verify DaisyUI classes are present in HTML output
    // Verify no custom CSS classes remain
}
```

**httptest Integration:**
```bash
# Test full handler stack
CGO_ENABLED=1 go test ./integration/prompts_integration_test.go
```

**Visual Regression:**
```bash
# Before migration
npm run screenshots

# After migration
npm run screenshots

# Compare images manually
```

### Final Verification

1. All Go tests pass: `CGO_ENABLED=1 go test ./... -v -race -cover`
2. Coverage maintained at 99%+ (no decrease from current 99.9%)
3. `templates/output.css` generated successfully and smaller than previous version
4. All pages render correctly in browser with DaisyUI theme applied
5. No custom CSS classes found in rendered HTML (grep search)
6. Visual screenshots match expected DaisyUI styling (differences acceptable per trade-offs)

## DaisyUI v5 Features Utilized

### Built-in Components Used
- **Cards**: `.card`, `.card-body`, `.card-title`, `.card-actions`
- **Buttons**: `.btn`, `.btn-primary`, `.btn-secondary`, `.btn-info`, `.btn-success`, `.btn-error`, `.btn-ghost`, `.btn-square`, `.btn-circle`
- **Inputs**: `.input`, `.input-bordered`, `.select`, `.select-bordered`, `.textarea`, `.textarea-bordered`, `.checkbox`, `.range`
- **Tables**: `.table`, `.table-zebra`
- **Badges**: `.badge`, `.badge-success`, `.badge-info`, `.badge-warning`, `.badge-error`, `.badge-outline`
- **Progress**: `.progress`
- **Alerts**: `.alert`, `.alert-info`, `.alert-warning`, `.alert-error`, `.alert-success`
- **Menus**: `.menu`, `.menu li`
- **Forms**: `.form-control`
- **Loading**: `.loading`, `.loading-spinner`, `.loading-bars`
- **Status**: `.status`, `.status-success`, `.status-error`
- **Buttons Group**: `.btn-group`

### Built-in Theme: `coffee`

The `coffee` theme provides:
- Warm, earthy tones matching our aesthetic
- Brown and cream color palette
- High contrast for readability
- Built-in hover and focus states

### Tailwind v4 Utilities Used

- **Animations**: `animate-spin`, `animate-ping`, `animate-pulse`, `animate-bounce`
- **Backdrop**: `backdrop-blur`, `backdrop-blur-sm`
- **Colors**: Arbitrary values `bg-[#hex]`, `text-[#hex]` for dynamic score theming
- **Layout**: `flex`, `flex-wrap`, `flex-col`, `grid`, `gap`, `p`, `m`, `w`, `h`
- **Typography**: `font-mono`, `font-sans`, `text-xs`, `text-sm`, `text-base`
- **Borders**: `border`, `rounded`, `shadow`
- **Transitions**: `transition`, `hover:`, `focus:` variants

## Migration Impact

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| **Custom CSS lines** | 1,065 | 0 (imports only) |
| **CSS complexity** | Custom gradients, keyframes, pseudo-elements | Zero custom CSS |
| **Component classes** | 25+ custom semantic classes | 0 custom classes |
| **Design tokens** | 12+ CSS variables | DaisyUI theme tokens |
| **Animations** | 3 custom keyframes | Tailwind built-ins |
| **Bundle size** | ~20KB custom CSS | ~15KB DaisyUI (tree-shaked) |
| **Maintenance burden** | High (custom CSS to maintain) | Low (DaisyUI handles complexity) |
| **Visual identity** | Unique glass + neon design | Standard DaisyUI cyberpunk theme |
| **Class complexity** | Semantic custom names | DaisyUI + Tailwind utilities |

## Key Design Decisions

### 1. Score Theming Strategy
**Decision:** Use Tailwind arbitrary values with dynamic color injection from JS

```html
<!-- Current implementation -->
<div class="score-cell score-20">20</div>

<!-- New implementation -->
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#ffa500] text-white shadow-sm">20</div>
```

**Rationale:** DaisyUI doesn't have semantic score colors. Arbitrary values allow full color control without custom CSS.

### 2. Body Background Strategy
**Decision:** Use DaisyUI `coffee` theme via `data-theme` attribute

```html
<html data-theme="coffee">
  <body class="bg-base-200 min-w-[1320px]">
    <!-- Content -->
  </body>
</html>
```

**Rationale:** Built-in themes are zero-maintenance and provide consistent color palettes. Coffee theme provides warm, earthy tones with good readability.

### 3. Glass Panel Strategy
**Decision:** Remove all glow effects, use standard DaisyUI `.card` components

```html
<!-- Current -->
<div class="glass-panel">...</div>

<!-- New -->
<div class="card card-compact bg-base-100 shadow-xl">...</div>
```

**Rationale:** Pseudo-element glows require custom CSS. DaisyUI `.card` has built-in shadows and borders.

### 4. Animation Strategy
**Decision:** Use Tailwind v4 built-in animations only

| Old Custom Animation | New Built-in | Context |
|-------------------|---------------|---------|
| `slowGlow` | `animate-pulse` | Panel effects |
| `shimmer` | `animate-pulse` | Heading effects |
| `pulse-connected` | `animate-ping` | Status indicators |
| Custom keyframes | Removed entirely | 0% custom CSS |

**Rationale:** Tailwind v4 includes comprehensive animation utilities. Custom animations require CSS keyframes.

### 5. Layout Strategy
**Decision:** Maintain original layout structure, replace custom classes with DaisyUI utilities

**Preserved:**
- `arena-shell` → `.flex`, `.grid`, `.gap`, `.p`
- `arena-topbar` → `.flex`, `.justify-between`, `.items-center`
- `arena-sidebar` → `.flex`, `.flex-col`
- `arena-main` → `.flex`, `.flex-col`, `.gap`

**Replaced:**
- `.glass-panel` → `.card`
- `.btn-gradient` → `.btn btn-primary`
- `.input-enhanced` → `.input input-bordered`

**Rationale:** Layout is structural. DaisyUI components replace styling, not structure.

## Configuration Files

### `tailwind.config.js`
```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./templates/**/*.html",
    "./tools/screenshots/**/*.html",
  ],
  plugins: [
    require('@tailwindcss/forms'),
    require('daisyui'),  // DaisyUI v5 plugin
  ],
  daisyui: {
    themes: [
      "coffee",     // Built-in warm, earthy theme
    ],
    darkTheme: "coffee",
  },
}
```

**Changes from previous:**
- Removed all custom color definitions (DaisyUI provides via themes)
- Removed custom spacing, shadows, fonts (DaisyUI provides defaults)
- Removed custom animations (Tailwind v4 has built-ins)
- Removed safelist (no longer needed - all classes are built-in)
- Added DaisyUI plugin
- Configured coffee theme

### `postcss.config.js`
```javascript
export default {
  plugins: {
    '@tailwindcss/postcss': {},  // Tailwind v4 syntax
    autoprefixer: {},
  },
}
```

**No changes:** Already compatible with DaisyUI v5.

### `templates/input.css`
```css
@import "tailwindcss";
@plugin "daisyui" {
  themes: coffee;
}
```

**Changes:** Deleted 1,065 lines of custom CSS. Only imports remain.

### `package.json`
```json
{
  "devDependencies": {
    "@tailwindcss/forms": "^0.5.11",
    "@tailwindcss/postcss": "^4.1.18",
    "autoprefixer": "^10.4.23",
    "playwright": "1.57.0",
    "postcss": "^8.5.6",
    "postcss-cli": "^11.0.1",
    "tailwindcss": "^4.1.18",
    "daisyui": "^5.0.0",  // ← ADDED
    "tdd-guard-jest": "^0.1.4",
    "tdd-guard-vitest": "^0.1.6"
  },
  "scripts": {
    "build:css": "postcss templates/input.css -o templates/output.css",
    "watch:css": "postcss templates/input.css -o templates/output.css --watch",
    "screenshots:install": "npx playwright install chromium",
    "screenshots": "npm run build:css && node tools/screenshots/capture.mjs"
  }
}
```

**Changes:** Added `daisyui` dependency. No script changes needed.

## Component Migration Reference

### Complete Class Mapping

| Custom Class | DaisyUI/Tailwind Equivalent | Notes |
|--------------|---------------------------|-------|
| **Layout** |
| `.arena-shell` | `flex flex-col min-h-screen bg-base-200 gap-3 p-3` | Tailwind utilities |
| `.arena-shell-solo` | `flex flex-col min-h-screen bg-base-200 p-3` | Tailwind utilities |
| `.arena-topbar` | `card bg-base-100 shadow-lg flex justify-between items-center gap-3 p-4` | DaisyUI + utilities |
| `.arena-topbar-solo` | `sticky top-0 z-50 w-full bg-base-100 shadow-lg p-4` | Tailwind utilities |
| `.arena-sidebar` | `card bg-base-100 shadow-lg flex flex-col gap-2 p-3 w-40` | DaisyUI + utilities |
| `.arena-main` | `flex-1 flex-col gap-3 overflow-auto` | Tailwind utilities |
| `.arena-main-solo` | `flex justify-center items-start flex-1` | Tailwind utilities |

| **Panels** |
| `.glass-panel` | `card bg-base-100 shadow-lg` | No glow effect |
| `.glass-panel-enhanced` | `card bg-base-100 shadow-xl` | No glow effect |
| `.glass-inset` | `card bg-base-200 shadow-inner` | DaisyUI + utilities |

| **Buttons** |
| `.btn-gradient` | `btn btn-primary` | Primary action buttons |
| `.btn-enhanced` | `btn btn-info` | Accent action buttons |
| `.btn-delete` | `btn btn-error` | Destructive actions |
| `.btn-icon` | `btn btn-square btn-ghost` | Icon-only buttons |
| `.scroll-button` | `btn btn-square btn-ghost` | Scroll navigation |

| **Inputs** |
| `.input-enhanced` | `input input-bordered w-full` | Form text inputs |
| `input[type="file"]` | `input input-bordered file-input w-full` | File uploads |
| `textarea` | `textarea textarea-bordered w-full` | Multi-line text |

| **Tables** |
| `table` | `table table-zebra w-full` | Data grids |
| `data-table` | `table table-zebra` | Same as above |

| **Badges & Status** |
| `.profile-badge` | `badge badge-success badge-outline` | Profile indicators |
| `.chip` | `badge badge-neutral` | Small labels |
| `.chip.pass` | `badge badge-success` | Pass status |
| `.chip.fail` | `badge badge-error` | Fail status |
| `.chip.tier` | `badge badge-info` | Tier badges |
| `.connection-status-connected` | `badge badge-success` | Connected state |
| `.connection-status-error` | `badge badge-error` | Error state |
| `.status-enhanced.connected` | `badge badge-success animate-ping` | Active with pulse |

| **Navigation** |
| `.nav-item` | `menu li` | Menu items |
| `.nav-item.is-active` | `menu li.active` | Active menu item |
| `.menu` | `menu bg-base-100` | Menu container |

| **Progress** |
| `.progress-bar-container` | `progress w-full` | Progress wrapper (DaisyUI handles) |
| `.progress-bar-standard-width` | `progress w-20` | Fixed width progress |
| `.progress-bar` | Removed (DaisyUI `.progress` has container) |

| **Score Cells** |
| `.score-cell-base` | `w-12 h-12 flex items-center justify-center font-bold rounded-lg` | Utilities only |
| `.score-cell-base:hover` | `hover:outline hover:outline-2 hover:outline-primary` | Tailwind utilities |
| `.score-0` | `bg-[#808080] text-white` | Arbitrary value |
| `.score-20` | `bg-[#ffa500] text-white` | Arbitrary value |
| `.score-40` | `bg-[#ffcc00] text-white` | Arbitrary value |
| `.score-60` | `bg-[#ffff00] text-white` | Arbitrary value |
| `.score-80` | `bg-[#ccff00] text-white` | Arbitrary value |
| `.score-100` | `bg-[#00ff00] text-white` | Arbitrary value |

| **Typography & Utilities** |
| `.mono` | `font-mono` | Tailwind built-in |
| `.margin-right-5` | `mr-1.25` | Tailwind built-in |
| `.hidden-data` | `hidden` | Tailwind built-in |
| `.sticky-header` | `sticky top-0` | Tailwind built-in |
| `.sticky-footer` | `sticky bottom-0` | Tailwind built-in |

| **Forms** |
| `.flex-form` | `form-control flex flex-wrap gap-2` | Tailwind utilities |
| `.filter-form` | `form-control flex flex-wrap gap-2` | Tailwind utilities |
| `.search-form` | `form-control flex gap-2` | Tailwind utilities |
| `.filter-container` | `div flex flex-wrap gap-2` | Tailwind utilities |
| `.file-import-form` | `form-control flex gap-2` | Tailwind utilities |
| `.results-management` | `form-control flex gap-2` | Tailwind utilities |

| **Layout Utilities** |
| `.title-row` | `flex items-center justify-between gap-3 p-4 flex-wrap` | Tailwind utilities |
| `.sticky-header` | `card-header` or `flex items-center justify-between p-4` | DaisyUI + utilities |
| `.sticky-footer` | `card-footer` or `flex items-center justify-between p-4` | DaisyUI + utilities |

| **Chart & Evaluation** |
| `.chart-container` | `div flex flex-col gap-2 overflow-y-auto` | Tailwind utilities |
| `.chart-wrapper` | `div relative w-full min-h-[340px]` | Tailwind utilities |
| `.evaluation-buttons` | `div flex flex-wrap gap-2 justify-center` | Tailwind utilities |
| `.score-buttons` | `div btn-group` | DaisyUI component |
| `.evaluation-form` | `form-control` | Tailwind utilities |

| **Alerts** |
| `.connection-status` | `badge badge-success` | DaisyUI component |
| `.connection-status-connected` | `badge badge-success animate-ping` | DaisyUI + animation |

## Migration Checklist

### Phase 1: Foundation
- [ ] Install DaisyUI v5 via npm
- [ ] Update `tailwind.config.js` with DaisyUI plugin and themes
- [ ] Verify `postcss.config.js` compatibility
- [ ] Delete all custom CSS from `templates/input.css` (keep imports only)
- [ ] Add `daisyui` to `package.json` devDependencies
- [ ] Run `npm run build:css` and verify output
- [ ] Run `CGO_ENABLED=1 go test ./... -v` - verify no regressions

### Phase 2: Component Migration (All Templates)
- [ ] Create class mapping documentation
- [ ] Update all templates to use DaisyUI components
- [ ] Remove all custom class references
- [ ] Verify Tailwind arbitrary values for dynamic colors
- [ ] Run full test suite with coverage

### Phase 3: Page Verification (Easiest → Hardest)

**Tier 1 - Simplest:**
- [ ] `nav.html` - Verify buttons and menu render
- [ ] `delete_prompt.html` - Verify form and buttons
- [ ] `delete_prompt_suite.html` - Verify form and buttons
- [ ] `delete_model.html` - Verify form and buttons
- [ ] `delete_profile.html` - Verify form and buttons
- [ ] `bulk_delete_prompts.html` - Verify form and buttons
- [ ] `confirm_refresh_results.html` - Verify form and buttons
- [ ] `import_prompts.html` - Verify form and buttons
- [ ] `import_results.html` - Verify form and buttons
- [ ] `import_error.html` - Verify alert component
- [ ] `move_prompt.html` - Verify form and buttons

**Tier 2 - Medium:**
- [ ] `new_prompt_suite.html` - Verify form with multiple inputs
- [ ] `edit_prompt_suite.html` - Verify form components
- [ ] `edit_prompt.html` - Verify form with textarea
- [ ] `edit_profile.html` - Verify form components
- [ ] `edit_model.html` - Verify form components
- [ ] `reset_prompts.html` - Verify form and danger button
- [ ] `reset_profiles.html` - Verify form and danger button
- [ ] `reset_results.html` - Verify form and danger button

**Tier 3 - Higher:**
- [ ] `profiles.html` - Verify tables, badges, progress, forms
- [ ] `prompt_list.html` - Verify tables, drag-drop, filtering
- [ ] `results.html` - Verify large tables, badges, status indicators
- [ ] `settings.html` - Verify all form types (inputs, selects, checkboxes, ranges)

**Tier 4 - Hardest:**
- [ ] `stats.html` - Verify charts render with DaisyUI colors
- [ ] `evaluate.html` - Verify complex score theming, evaluation forms
- [ ] `design_preview.html` - Verify layout with multiple cards

### Phase 4: Final Verification
- [ ] Take baseline screenshots (before migration)
- [ ] Run complete test suite: `CGO_ENABLED=1 go test ./... -v -race -cover`
- [ ] Take post-migration screenshots
- [ ] Manually compare screenshots (document acceptable differences per trade-offs)
- [ ] Verify coverage maintained at 99%+
- [ ] Check all templates have no custom CSS classes
- [ ] Verify `templates/output.css` size is smaller than before
- [ ] Cross-browser testing (Chrome, Firefox, Safari, Edge)
- [ ] Remove any unused temporary files

## Success Metrics

Upon completion, we will have achieved:

1. **0 lines of custom CSS** (except Tailwind/DaisyUI imports)
2. **100% DaisyUI v5 + Tailwind v4** styling
3. **All tests passing** with coverage ≥ 99%
4. **Visual consistency** across all pages using DaisyUI coffee theme
5. **Maintainable codebase** using industry-standard tools
6. **No CSS maintenance burden** (DaisyUI handles complexity)
7. **Zero pseudo-elements** (no `::before` or `::after` for effects)
8. **Zero custom keyframes** (only Tailwind built-in animations)
9. **Zero custom CSS variables** (DaisyUI theme tokens instead)
10. **Complete SSR flow verified** via httptest integration tests

## Rollback Plan

If migration fails or visual results are unacceptable:

1. Revert `tailwind.config.js` to previous version
2. Revert `templates/input.css` from git (full custom CSS restored)
3. Revert all template changes to use custom classes
4. Remove `daisyui` from `package.json`
5. Run `npm run build:css` to regenerate output.css
6. Run full test suite to verify rollback success

## Timeline Estimate

- **Phase 1 (Foundation)**: 1-2 hours
- **Phase 2 (Component Migration)**: 4-6 hours (24 templates)
- **Phase 3 (Page Verification)**: 2-3 hours (httptest testing)
- **Phase 4 (Final Verification)**: 2-3 hours (screenshots, cross-browser)
- **Total**: 9-14 hours of focused work

## References

- DaisyUI v5 Documentation: https://daisyui.com/docs/v5/
- Tailwind CSS v4 Documentation: https://tailwindcss.com/
- Tailwind v4 Release Notes: https://tailwindcss.com/blog/tailwindcss-v4-alpha
- DaisyUI GitHub: https://github.com/saadeghi/daisyui
- Tailwind CSS GitHub: https://github.com/tailwindlabs/tailwindcss
