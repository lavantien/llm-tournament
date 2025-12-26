# LLM Tournament Arena — Design Concept (Tailwind v4 + DaisyUI v5 Edition)

## Goal

Transform UI to **100% pure Tailwind v4 + DaisyUI v5** - zero custom CSS, built-in components only. Deliver a clean, maintainable dashboard using industry-standard tools while maintaining SSR Go template architecture.

## Constraints (Non-Negotiable)

- **ZERO custom CSS** beyond Tailwind/DaisyUI imports
- No custom CSS variables, keyframes, or layer directives
- No CSS pseudo-elements (`::before`, `::after`) for visual effects
- No complex multi-layer gradients in CSS (use DaisyUI themes or Tailwind arbitrary values)
- DaisyUI built-in themes only (no custom theme definitions)
- Tailwind built-in animations only (no custom `@keyframes`)
- Desktop-first optimization maintained (from original design concept)
- SSR architecture preserved (`html/template` package)
- Testing verification through Go template rendering with `httptest` (not CSS unit tests)

## Visual Direction

**"Neon Glass Foundry" → "Clean Coffee"**

- **Base**: DaisyUI `coffee` theme (warm, earthy tones with good contrast)
- **Panels**: DaisyUI `.card` components (no glass glow effects)
- **Buttons**: DaisyUI `.btn` variants (`.btn-primary`, `.btn-info`, `.btn-error`, `.btn-ghost`, `.btn-square`)
- **Inputs**: DaisyUI `.input`, `.select`, `.textarea`, `.checkbox` components
- **Tables**: DaisyUI `.table`, `.table-zebra` with built-in styling
- **Badges**: DaisyUI `.badge`, `.badge-success`, `.badge-info`, `.badge-warning`, `.badge-error`
- **Animations**: Tailwind v4 built-in (`animate-spin`, `animate-ping`, `animate-pulse`, `animate-bounce`)
- **Typography**: Tailwind `font-sans` (Inter) and `font-mono` (JetBrains Mono)

## Trade-offs Accepted (Zero Custom CSS)

| Original Feature | Built-in Solution | Trade-off |
|-----------------|-------------------|-----------|
| **Glass panel glow effects** | DaisyUI `.card` with `shadow-lg` | No glow overlay, cleaner appearance |
| **3-layer body background gradients** | DaisyUI `cyberpunk` theme | Different gradient pattern, similar dark aesthetic |
| **Grid overlay texture** | Removed entirely | No grid texture, clean background |
| **Custom animations** (`slowGlow`, `shimmer`, `pulse-connected`) | Tailwind `animate-pulse`, `animate-ping`, `animate-spin` | Simpler animations, less dramatic |
| **Complex gradient borders** (cyan→magenta→violet) | DaisyUI `.btn-primary`, `.btn-info`, `.btn-accent` | Single color per button variant |
| **Custom button hover effects** | DaisyUI `.btn` with built-in hover states | Simpler hover transitions |
| **Dynamic score colors** (CSS variables) | Tailwind arbitrary values `bg-[#color]` | More verbose HTML but flexible |
| **Pseudo-element glows** | Removed entirely | 0% custom CSS requirement |

## Typography

- **UI**: **Inter** (clean, modern, highly legible at dense sizes)
- **Data/Code**: **JetBrains Mono** (scores, tiers, IDs, logs, timestamps)
- **Fallbacks**: `system-ui, Segoe UI, Roboto, Arial, sans-serif` and `ui-monospace, Consolas, Menlo, monospace`
- **Sizing**: Same as original (`text-xs`, `text-sm`, `text-base`)

## Color Palette

### DaisyUI Coffee Theme

DaisyUI `coffee` theme provides:
- Warm, earthy tones with good contrast
- Brown and cream color palette
- High contrast for readability
- Built-in hover and focus states
- Semantic color names (`primary`, `secondary`, `accent`, `success`, `info`, `warning`, `error`)

### Score Theming (Arbitrary Values)

Dynamic score cells use Tailwind arbitrary values:
```html
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#ffa500] text-white shadow-sm">20</div>
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#ffcc00] text-white shadow-sm">40</div>
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#ffff00] text-white shadow-sm">60</div>
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#ccff00] text-white shadow-sm">80</div>
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[#00ff00] text-white shadow-sm">100</div>
```

## Layout Structure

### Preserved from Original

- **Top Bar**: Product identity + suite context + live connection state
- **Left Rail**: Dense navigation (Results / Stats / Prompts / Profiles / Evaluate / Settings), quick actions, "hotkeys" hints
- **Main Grid**: 2-column dashboard (table-heavy left; charts/insights right)

### Updated with DaisyUI + Tailwind

- **Layout Wrappers**: `flex`, `grid`, `gap`, `p`, `m` utilities
- **Panels**: `.card` with DaisyUI styling
- **Forms**: `.input input-bordered`, `.select select-bordered`, `.btn`, `.btn-primary`
- **Tables**: `.table`, `.table-zebra` for data grids
- **Status Indicators**: `.badge`, `.badge-success`, `.badge-error`, `.animate-ping`

## UI Elements

### DaisyUI v5 Components

#### Cards & Panels

- `.card` - Glass panel replacement (no glow)
- `.card-body` - Content wrapper
- `.card-title` - Heading styling
- `.card-actions` - Action buttons wrapper
- `shadow-lg`, `shadow-xl` - Built-in shadows

### Buttons

- `.btn` - Base button
- `.btn-primary` - Primary actions (cyan accent)
- `.btn-info` - Secondary actions (magenta/violet accent)
- `.btn-success` - Success actions
- `.btn-error` - Danger/destructive actions
- `.btn-ghost` - Low emphasis actions
- `.btn-square`, `.btn-circle` - Icon-only buttons
- `hover:`, `focus:` - Built-in state variants
- `animate-spin`, `animate-ping` - Loading/active states

### Inputs & Forms

- `.input` - Text inputs
- `.input-bordered` - Text inputs with borders
- `.select` - Dropdown selects
- `.select-bordered` - Selects with borders
- `.textarea` - Multi-line text
- `.textarea-bordered` - Textareas with borders
- `.checkbox` - Checkboxes
- `.checkbox-sm` - Small checkboxes
- `.range` - Range sliders
- `.file-input` - File uploads

### Tables & Data

- `.table` - Base table
- `.table-zebra` - Zebra striping
- `.table-hover` - Row hover effects
- `.badge` - Small labels/status
- `.badge-success`, `.badge-info`, `.badge-warning`, `.badge-error` - Color variants
- `.badge-outline` - Outline variant

### Navigation

- `.menu` - Menu container
- `.menu li` - Menu items
- `.menu li.active` - Active state
- `hover:` - Hover effects (built-in)

### Progress & Status

- `.progress` - Progress bar
- `.progress-bar` - Progress fill (deprecated, use `.progress`)
- `.status` - Status indicator
- `.status-success`, `.status-error` - Color variants
- `animate-ping` - Pulse animation for active status

### Alerts

- `.alert` - Alert container
- `.alert-info`, `.alert-success`, `.alert-warning`, `.alert-error` - Color variants

### Loading

- `.loading` - Loading wrapper
- `.loading-spinner` - Spinner animation
- `.loading-bars` - Bars animation
- `.loading-dots` - Dots animation
- `.loading-infinity` - Infinity animation

## Animations (Tailwind v4 Built-in Only)

- `animate-spin` - Rotation (loading spinners)
- `animate-ping` - Scale/fade pulse (status indicators)
- `animate-pulse` - Opacity fade (skeleton loaders, panel effects)
- `animate-bounce` - Bounce (scroll indicators)
- `transition` - All transition properties
- `duration-*` - Animation duration utilities
- `ease-*` - Timing functions
- `delay-*` - Transition delays

## Responsive Design

- **Desktop-first**: Optimize for large screens + information density
- **Mobile**: DaisyUI provides responsive utilities, but we prioritize desktop
- **Breakpoints**: Tailwind defaults (`sm:`, `md:`, `lg:`, `xl:`)
- **Container**: `max-w-7xl mx-auto` or arbitrary `max-w-[1320px]`

## Accessibility

- DaisyUI components include ARIA attributes by default
- Focus states built-in (`focus:` variants)
- Semantic HTML structure maintained
- Keyboard navigation support
- Screen reader friendly color contrasts

## Performance

- **Tree-shaking**: DaisyUI v5 removes unused component styles
- **Smaller bundle**: ~15KB vs 1,065 lines custom CSS
- **No runtime JS**: All styling is CSS-only
- **Optimized rendering**: Tailwind JIT compiler
- **CSS caching**: DaisyUI components cache efficiently

## Testing Methodology

### Go SSR Testing

All templates are Go `html/template` files. We verify:

1. **Template rendering**: Go handlers render templates with test data
2. **DaisyUI classes present**: Verify correct class names in HTML output
3. **No custom CSS classes**: Ensure all custom classes replaced
4. **Arbitrary values correct**: Dynamic score colors use correct hex values
5. **Theme attribute applied**: Verify `data-theme="cyberpunk"` in `<html>` tag

### Integration Tests

- `integration/prompts_integration_test.go` - Full SSR flow with `cmp`
- `httptest` - HTTP handler testing with rendered HTML
- Coverage targets: `templates/` directory at 100%

### Visual Regression

- **Baseline screenshots**: Capture all pages before migration
- **Post-migration screenshots**: Compare with new design
- **Acceptable differences**: Theme changes, missing effects (per trade-offs)
- **Unacceptable regressions**: Layout breaks, missing components, broken functionality

## Libraries (CDN Only)

### Required

- Tailwind CSS v4.1.18 (already installed)
- DaisyUI v5.0.0 (will install)
- `@tailwindcss/forms` v0.5.11 (already installed)
- `@tailwindcss/postcss` v4.1.18 (already installed)
- PostCSS v8.5.6 (already installed)
- Autoprefixer v10.4.23 (already installed)

### Optional

- Font Awesome 6 (icons, if needed)
- Playwright 1.57.0 (screenshot testing)

## Implementation Notes

### Configuration Files

**`tailwind.config.js`**: Minimal, DaisyUI plugin only
**`postcss.config.js`**: Unchanged (v4 compatible)
**`templates/input.css`**: Only Tailwind/DaisyUI imports, zero custom CSS

### Template Migration Strategy

1. **Replace custom classes with DaisyUI components**
2. **Use Tailwind utilities for layout and spacing**
3. **Use Tailwind arbitrary values for dynamic colors**
4. **Maintain Go template structure** (no syntax changes)
5. **Test with real data via integration tests**

### Migration Phases

**Phase 1**: Foundation setup (DaisyUI install, config updates)
**Phase 2**: Component migration (all 24 templates at once)
**Phase 3**: Page verification (httptest testing, easiest → hardest)
**Phase 4**: Final verification (screenshots, cross-browser, cleanup)

### Migration Order (Easiest → Hardest)

**Tier 1 - Simplest**:
- `nav.html` - Buttons and menu
- `delete_prompt.html` - Form with confirm
- `delete_prompt_suite.html` - Form with confirm
- `delete_model.html` - Form with confirm
- `delete_profile.html` - Form with confirm
- `bulk_delete_prompts.html` - Form with confirm
- `confirm_refresh_results.html` - Form with confirm
- `import_prompts.html`, `import_results.html`, `move_prompt.html` - Forms

**Tier 2 - Medium**:
- `new_prompt_suite.html` - Form with multiple inputs
- `edit_prompt_suite.html` - Form components
- `edit_prompt.html` - Form with textarea
- `edit_profile.html` - Form components
- `edit_model.html` - Form components
- `reset_prompts.html` - Form with danger button
- `reset_profiles.html` - Form with danger button
- `reset_results.html` - Form with danger button
- `import_error.html` - Alert component

**Tier 3 - Higher**:
- `profiles.html` - Tables, forms, badges, progress
- `prompt_list.html` - Tables, drag-drop, filtering
- `results.html` - Large tables, badges, status
- `settings.html` - All form types

**Tier 4 - Hardest**:
- `stats.html` - Charts with DaisyUI colors
- `evaluate.html` - Complex score theming, evaluation forms
- `design_preview.html` - Layout with multiple panels

## Key Design Decisions

### 1. Zero Custom CSS Rule

**Decision**: All styling must use Tailwind v4 utilities or DaisyUI v5 components. No custom CSS allowed.

**Implementation**:
- Delete 1,065 lines from `templates/input.css`
- Keep only imports: `@import "tailwindcss";` and `@plugin "daisyui" { themes: ... };`
- No `@layer base`, `@layer components`, `@layer utilities` with custom styles
- No `@keyframes` for custom animations
- No CSS variables (`--custom-var`)
- No pseudo-elements for visual effects

### 2. Score Theming Strategy

**Decision**: Use Tailwind arbitrary values with dynamic color injection from Go templates

**Implementation**:
```html
<!-- Go template provides score value -->
<div class="w-12 h-12 flex items-center justify-center font-bold rounded-lg bg-[{{.ScoreColor}}] text-white shadow-sm">
    {{.Score}}
</div>
```

**Rationale**: DaisyUI doesn't have semantic score colors. Arbitrary values allow full color control without custom CSS.

### 3. Body Background Strategy

**Decision**: Use DaisyUI `coffee` theme via `data-theme` attribute

**Implementation**:
```html
<html data-theme="coffee">
  <body class="bg-base-200 min-w-[1320px]">
    <!-- Content -->
  </body>
</html>
```

**Rationale**: Built-in themes are zero-maintenance and provide consistent color palettes. Coffee theme provides warm, earthy tones with good readability.

### 4. Glass Panel Strategy

**Decision**: Remove all glow effects, use standard DaisyUI `.card` components

**Implementation**:
```html
<!-- Original (removed) -->
<div class="glass-panel">...</div>

<!-- New -->
<div class="card card-compact bg-base-100 shadow-lg">...</div>
```

**Rationale**: Pseudo-element glows require custom CSS. DaisyUI `.card` has built-in shadows and borders.

### 5. Animation Strategy

**Decision**: Use Tailwind v4 built-in animations only

**Mapping**:
| Old Custom Animation | New Built-in Animation | Context |
|-------------------|---------------------|---------|
| `slowGlow` | `animate-pulse` | Panel effects |
| `shimmer` | `animate-pulse` | Heading effects |
| `pulse-connected` | `animate-ping` | Status indicators |
| Custom keyframes | Removed entirely | 0% custom CSS |

**Rationale**: Tailwind v4 includes comprehensive animation utilities. Custom animations require CSS keyframes.

### 6. Layout Strategy

**Decision**: Maintain original layout structure, replace custom classes with DaisyUI utilities

**Preserved**:
- `arena-shell` → `flex`, `grid`, `gap`, `p` utilities
- `arena-topbar` → `.flex`, `.justify-between`, `.items-center`
- `arena-sidebar` → `.flex`, `.flex-col`
- `arena-main` → `.flex`, `.flex-col`, `.gap`

**Replaced**:
- `.glass-panel` → `.card`
- `.btn-gradient` → `.btn btn-primary`
- `.input-enhanced` → `.input input-bordered`

**Rationale**: Layout is structural. DaisyUI components replace styling, not structure.

## Migration Impact

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| **Custom CSS lines** | 1,065 | 0 (imports only) | -100% (success) |
| **CSS complexity** | Custom gradients, keyframes, pseudo-elements | Zero custom CSS | Total elimination |
| **Component classes** | 25+ custom semantic classes | 0 custom classes | DaisyUI + Tailwind |
| **Design tokens** | 12+ CSS variables | DaisyUI theme tokens | Industry standard |
| **Animations** | 3 custom keyframes | Tailwind built-ins | Better performance |
| **Bundle size** | ~20KB custom CSS | ~15KB DaisyUI (tree-shaked) | -25% reduction |
| **Maintenance burden** | High (custom CSS) | Low (DaisyUI) | Significant improvement |
| **Visual identity** | Unique glass + neon design | Standard DaisyUI cyberpunk | Clean, consistent |
| **Class complexity** | Semantic custom names | DaisyUI + Tailwind utilities | More verbose, industry-standard |

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
9. **Zero custom CSS variables** (DaisyUI theme tokens)
10. **Complete SSR flow verified** via httptest integration tests
11. **Desktop-first optimization** maintained (from original design concept)

## Risks & Mitigations

### Risk: Visual Changes (Medium)

**Risk**: Significant visual changes from glass glow → standard DaisyUI cards

**Mitigation**:
- Accept trade-offs documented in requirements contract
- Use `shadow-lg`, `shadow-xl` for depth
- Maintain consistent aesthetic with coffee theme
- User acceptance testing before finalizing

### Risk: Learning Curve (Low)

**Risk**: Team needs to learn DaisyUI component API

**Mitigation**:
- Comprehensive component mapping documentation
- DaisyUI docs available at https://daisyui.com/docs/v5/
- Built-in Tailwind utilities unchanged

### Risk: Testing Complexity (Low)

**Risk**: Testing 24 templates with Go SSR + httptest

**Mitigation**:
- Automated integration tests
- Visual regression screenshots
- Incremental page verification (easiest → hardest)

### Risk: Bundle Size (Low)

**Risk**: DaisyUI adds ~15KB vs ~20KB custom CSS

**Mitigation**:
- Tree-shaking removes unused component styles
- Overall size decrease from custom CSS
- Only use needed DaisyUI components

## Rollback Plan

If migration fails or visual results are unacceptable:

1. **Revert `tailwind.config.js`** to previous version from git
2. **Revert `templates/input.css`** from git (full custom CSS restored)
3. **Revert all template changes** to use custom classes via git
4. **Remove `daisyui` from `package.json` devDependencies
5. **Run `npm run build:css`** to regenerate output.css
6. **Run full test suite** to verify rollback success

## References

- DaisyUI v5 Documentation: https://daisyui.com/docs/v5/
- Tailwind CSS v4 Documentation: https://tailwindcss.com/
- Tailwind v4 Release Notes: https://tailwindcss.com/blog/tailwindcss-v4-alpha
- DaisyUI GitHub: https://github.com/saadeghi/daisyui
- Tailwind CSS GitHub: https://github.com/tailwindlabs/tailwindcss

## Next Steps

1. Review and approve this design concept
2. Review and approve migration plan (DESIGN_ROLLOUT.md)
3. Begin Phase 1: Foundation setup
4. Execute Phase 2: Component migration (all templates)
5. Execute Phase 3: Page verification (httptest testing)
6. Execute Phase 4: Final verification (screenshots, cross-browser)
7. Update README.md with completed migration status
