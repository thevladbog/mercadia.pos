# UI Components and Theme System

Mercadia frontends share interactive UI through `@mercadia/ui` (`frontend/packages/ui`).

Design references live under [`docs/Design/Design/`](../Design/Design/):

| Folder | Application | Notes |
|--------|-------------|-------|
| [`Админка/`](../Design/Design/Админка/) | `admin-web` | Sidebar AppShell; light/dark toggle with persist |
| Root `E _ *.png` | `pos-terminal` | Sale/return accent from layout template |
| [`КСО/`](../Design/Design/КСО/) | `sco-terminal` | SCO accent preset |
| [`Старший кассир/`](../Design/Design/Старший кассир/) | `senior-cashier-terminal` | Dark surface |

## Stack

- **Radix UI** for accessible primitives (Dialog, Tabs, Label, Slot).
- **CSS custom properties** for theming — no Tailwind, no shadcn.
- **class-variance-authority** for component variants.

Apps import styles once:

```ts
import '@mercadia/ui/styles.css';
```

Wrap the app root with `ThemeProvider`:

```tsx
import { ThemeProvider } from '@mercadia/ui';

<ThemeProvider
  defaultTheme={{ surface: 'admin', colorMode: 'light', accentPreset: 'neutral' }}
  persist
>
  {children}
</ThemeProvider>
```

## Three-layer token model

### Layer 1 — Primitives (`primitives.css`)

Raw palette values (`--primitive-*`). **Components must not reference these directly.**

### Layer 2 — Semantic tokens (`semantic-light.css`, `semantic-dark.css`)

Components and app layout utilities use **`--ui-*`** only:

- `--ui-bg`, `--ui-surface`, `--ui-text`, `--ui-border`
- `--ui-accent`, `--ui-accent-hover`, `--ui-accent-muted`, `--ui-accent-foreground`
- `--ui-success`, `--ui-warning`, `--ui-danger`, `--ui-info` (+ muted variants)

Active color mode is selected via `data-color-mode="light|dark"` on `<html>`.

### Layer 3 — Surface presets + runtime accent

**Surface** (`data-surface`) adds layout/density tokens without changing accent logic:

- `admin`, `terminal`, `sco`, `senior-cashier`

**Runtime accent** is applied by `applyTheme()` / `ThemeProvider`:

```tsx
import { applyTheme } from '@mercadia/ui';

applyTheme({
  surface: 'terminal',
  colorMode: 'light',
  accentPreset: 'return', // or accent: '#2563EB' from layout template API
});
```

This sets inline CSS variables on `<html>`:

```html
<html data-surface="terminal" data-color-mode="light" style="--ui-accent: #2563EB; ...">
```

## Accent presets

Built-in presets in `presets.ts` (extend the record, no new CSS file):

| Preset | Use case | Default hex |
|--------|----------|-------------|
| `sale` | Sale register | `#FF6600` |
| `return` | Return register | `#2563EB` |
| `sco` | Self-checkout | `#F25F1C` |
| `neutral` | Admin default | `#FF6600` |

Layout templates and Color Schemes may supply `accentPreset` or `accentColor`; POS terminals call `applyTheme` when binding a template at shift start.

## Components (v1)

| Component | Notes |
|-----------|-------|
| `Button` | `primary` uses `--ui-accent`; `secondary`, `ghost`, `link` |
| `Badge` | Semantic + `accent` variant |
| `Dialog`, `DetailDialog`, `FormDialog` | Radix dialog; `DetailDialog` read-only; `FormDialog` submit forms (Safe/EoD) |
| `Input`, `Select`, `Textarea`, `Checkbox`, `Label`, `Field` | Form controls |
| `Tabs`, `PillTabs` | Radix tabs |
| `Card`, `CardHeading` | Panel/card layout |
| `LayoutGrid` | Product tile grid from layout template JSON |
| `LayoutGridEditor` (admin-web) | Controlled editor for layout template grid JSON on create/edit forms |
| `Numpad`, `Stepper` | Touch kiosk controls |
| `ThemePreview` | Scoped accent preview without mutating global admin theme |

New interactive controls belong in `@mercadia/ui`, not as global CSS in apps.

## Storybook

`@mercadia/ui` includes a Storybook catalog for shared components and design tokens.

```bash
cd frontend
pnpm storybook:ui
pnpm build-storybook:ui
```

Use the Storybook toolbar to switch `surface`, `colorMode`, and `accentPreset`. The Foundations
stories show computed CSS custom properties for primitive colors, semantic colors, spacing,
radii, shadows, and surface sizing tokens.

## Central backend branding APIs

Color schemes and layout templates are stored in central-backend:

| Resource | Endpoints |
|----------|-----------|
| Color schemes | `GET/POST /v1/color-schemes`, `GET/PATCH /v1/color-schemes/{schemeId}` |
| Layout templates | `GET/POST /v1/layout-templates`, `GET/PATCH /v1/layout-templates/{templateId}` |

Layout template responses include `resolvedAccentPreset` and `resolvedAccentColor` for POS/SCO clients.

Accent resolution order: template `accentColor` → template `accentPreset` → linked color scheme → default by `kind` (`sale` / `return` / `sco`).

Admin CRUD lives under `/central/color-schemes` and `/central/layout-templates` (central admin only). Reads require central session or sync API key with reporting permission.

## POS terminal scaffold

`frontend/apps/pos-terminal` loads a template via `?templateId=` or `VITE_LAYOUT_TEMPLATE_ID`, calls `applyTheme` from resolved accent fields, and renders `LayoutGrid` + `Numpad` demo. It also includes a Store Edge checkout demo that prepares an operational day/shift, opens a receipt, scans a product, captures a mock payment, and creates a mock fiscal document. Set `VITE_CENTRAL_SESSION_TOKEN` for central layout-template access and `VITE_STORE_EDGE_URL` only when bypassing the Vite proxy.

```bash
cd frontend
pnpm --filter pos-terminal dev
# open http://localhost:5174/?templateId=sale-standard
```

## Adding a new theme dimension

1. **New surface** — add `surfaces/foo.css`, extend `Surface` in `types.ts`, document design folder mapping.
2. **New accent preset** — add entry to `ACCENT_PRESETS` in `presets.ts`.
3. **Custom accent** — pass `accent: '#hex'` in `ThemeConfig` (template/API).
4. **New color mode** — add `semantic-*.css` and wire `ColorMode` (v1: `light` + `dark` only).

## Admin-web migration

- `ThemeProvider` in `Root.tsx` with `surface: 'admin'`, `colorMode: 'light'`, `accentPreset: 'neutral'`, and `persist` (localStorage key `mercadia-ui-theme`).
- **AppShell** — sidebar navigation in `AppSidebar.tsx` (`AppLayout.tsx`); grouped Central / Store ops / Admin links with `NavLink` active state.
- **Dark mode** — `ThemeToggle` in the top bar calls `useTheme().setTheme({ ...theme, colorMode })`; shell chrome (`app-sidebar`, `app-header`, panels, tables) uses `--ui-*` tokens so light/dark switches apply without extra CSS files.
- Prefer `@mercadia/ui` components over native `<button>` with global CSS.
- Legacy native buttons remain styled via token-based compat rules in `index.css` until pages migrate.
- Cash/EoD form modals use `FormDialog` directly from `@mercadia/ui`; read-only drill-down modals use `DetailDialog`. Legacy `CashModal` / `DetailModal` wrappers and `.modal-backdrop` CSS removed.
- Layout template create/edit forms use `LayoutGridEditor` (categories, icon URLs, product validation on publish).
- Admin pages use `@mercadia/ui` `Button`; legacy native button compat rules removed from `index.css`.

## Verification

```bash
cd frontend
pnpm --filter @mercadia/ui test
pnpm --filter @mercadia/ui typecheck
pnpm verify
```
