# Admin Web i18n

Canonical internationalization rules for [`frontend/apps/admin-web`](../../frontend/apps/admin-web).

## Stack

- `i18next` + `react-i18next`
- Locale files: [`frontend/apps/admin-web/src/i18n/locales/ru.json`](../../frontend/apps/admin-web/src/i18n/locales/ru.json) (primary), [`en.json`](../../frontend/apps/admin-web/src/i18n/locales/en.json) (fallback)
- Config: [`frontend/apps/admin-web/src/i18n/config.ts`](../../frontend/apps/admin-web/src/i18n/config.ts)
- Bootstrap: `initI18n()` in [`main.tsx`](../../frontend/apps/admin-web/src/main.tsx); `I18nextProvider` in [`Root.tsx`](../../frontend/apps/admin-web/src/Root.tsx)

Defaults:

- Default locale: `ru`
- Fallback locale: `en`
- Persisted in `localStorage` under key `mercadia.admin.locale`

## Usage pattern

```tsx
import { useTranslation } from 'react-i18next';

export function ExamplePage() {
  const { t } = useTranslation();

  return <h2>{t('monitoring.title')}</h2>;
}
```

## Key rules

1. **Every user-facing string** in admin-web TSX must use `t()` — labels, buttons, empty states, validation messages, success notices.
2. **Add keys to both locale files** in the same PR (`ru.json` and `en.json`).
3. **Key naming:** nested by domain — `nav.*`, `common.*`, `safe.*`, `eod.*`, `monitoring.*`, etc.
4. **Do not translate:**
   - Route paths
   - Orval-generated types
   - Raw API enum values (`status`, `kind`, movement `type`) unless explicit display-label keys are added
5. **Backend errors:** use `getApiErrorMessage()` and show server `detail` as-is; only wrap UI chrome in i18n.
6. **Formatting:** use [`reporting-utils.ts`](../../frontend/apps/admin-web/src/pages/reporting-utils.ts) `formatMinorAmount()` and `formatTimestamp()` — they read `i18n.language` for `Intl` locale (`ru-RU` / `en-US`).
7. **Language switcher:** [`LanguageSwitcher.tsx`](../../frontend/apps/admin-web/src/components/LanguageSwitcher.tsx) in the app header.

## Store-edge integration (admin-web)

When calling store-edge **command** endpoints from admin-web:

### Vite dev proxy

Store-edge paths must be proxied to `:8081` **before** the central catch-all `/v1` rule in [`vite.config.ts`](../../frontend/apps/admin-web/vite.config.ts):

```typescript
'^/v1/stores/[^/]+/(monitoring|terminals|cash-|bank-|business-|operation-journal|operational-days|shifts)'
```

Also proxy `/v1/operational-days`, `/v1/shifts`, `/v1/receipts`, `/v1/terminals`.

Optional env override: `VITE_STORE_EDGE_URL`.

### Orval clients

Import hooks/functions from [`@mercadia/api-clients-store-edge`](../../frontend/packages/api-clients/store-edge). Relevant tags:

- `cash-office` — balances, movements, recounts, bank collection, expenses
- `store-operations` — operational day, shifts, journal
- `checkout` — receipt read for EoD blocker drill-down
- `monitoring` — terminal KPIs and SSE

### Idempotency

All store-edge command endpoints require an `Idempotency-Key` header. Use:

```typescript
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';

await createCashMovement(storeId, body, { headers: createIdempotencyHeaders() });
```

Pattern reference: [`RegisterStorePage.tsx`](../../frontend/apps/admin-web/src/pages/RegisterStorePage.tsx).

After a successful `202`, invalidate affected React Query keys (see `invalidateSafeQueries()` in [`cash-mutation-utils.ts`](../../frontend/apps/admin-web/src/pages/cash-mutation-utils.ts) and `invalidateEodQueries()` in [`eod-mutation-utils.ts`](../../frontend/apps/admin-web/src/pages/eod-mutation-utils.ts)).

### Cash operations UI

- Admin panel directly manages the **safe** (see [`docs/open-questions.md`](../open-questions.md)).
- Forms collect store-edge `actorId` and `approvedById` explicitly — there is no actors list API yet. Demo seed actors: `admin-1`, `senior-1`, `cashier-1`.
- **Separation of duties:** `actorId` must differ from `approvedById` when both are required. Enforced server-side; validate client-side too.
- Write UI is gated to `central_admin` in the frontend; store-edge enforces business rules independently.

### EoD close command

Close operational day from [`StoreEodPage.tsx`](../../frontend/apps/admin-web/src/pages/StoreEodPage.tsx) via `closeOperationalDay()`:

```typescript
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { invalidateEodQueries } from '@/pages/eod-mutation-utils.js';

await closeOperationalDay(
  operationalDayId,
  {
    closedById: userId,
    overrideNoSales: true,
    overrideActorId: 'admin-1',
  },
  { headers: createIdempotencyHeaders() },
);

await invalidateEodQueries(queryClient, storeId, operationalDayId);
```

When the only blocker is `no_sales_receipts` (`requires_admin_override`), send `overrideNoSales: true` and a distinct `overrideActorId`. Hard blockers (`open_cashier_shift`, etc.) must be resolved before close succeeds.

### EoD open command

Open operational day from the no-open-day empty state via `openOperationalDay()`:

```typescript
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { invalidateEodAfterOpen, todayBusinessDate } from '@/pages/eod-mutation-utils.js';

const response = await openOperationalDay(
  {
    storeId,
    businessDate: todayBusinessDate(),
    openedById: userId,
  },
  { headers: createIdempotencyHeaders() },
);

if (response.status === 202) {
  await invalidateEodAfterOpen(queryClient, storeId, response.data.operationalDay.id);
}
```

Returns **409** if the store already has an open day.

### EoD shift close command

Close cashier shift from the Open Shifts tab via `closeShift()`:

```typescript
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { invalidateShiftCloseQueries } from '@/pages/shift-mutation-utils.js';

await closeShift(
  shiftId,
  {
    closingCashMinor: 0,
  },
  { headers: createIdempotencyHeaders() },
);

// When closingCashMinor > 0, also send safeId, actorId, approvedById (distinct actors).

await invalidateShiftCloseQueries(queryClient, storeId, operationalDayId);
```

Returns **409** `shift_close_blocked` when the shift has unresolved receipts.

## Checklist for admin-web PRs

- [ ] No new hardcoded user-facing strings in TSX
- [ ] New keys added to `ru.json` and `en.json`
- [ ] Store-edge mutations send `Idempotency-Key`
- [ ] Query cache invalidated after successful commands
- [ ] Vite proxy updated if new store-edge path prefixes are introduced
