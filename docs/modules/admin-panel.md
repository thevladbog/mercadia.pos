# Admin Panel Specification

## Purpose

The admin panel is the operational control plane for Mercadia POS. It combines real-time
monitoring, cash office management, shift close, reconciliation, master data, integrations,
device setup, and UI/template configuration.

## Navigation

The design shows these main areas:

- Monitoring.
- Safe.
- EoD.
- Catalog and accounting.
- Products.
- Reports.
- Reconciliation.
- Management.
- Users.
- POS and SCO.
- Layout templates.
- Receipt templates.
- Color schemes.
- Payment methods.
- System.
- Settings.
- Integrations.
- Event log.

The admin panel should use role-aware navigation. Users should only see sections they can
access or should see disabled sections with a clear permission reason, depending on policy.

## Real-Time Monitoring

Monitoring gives a live view of all POS and SCO nodes in the store.

Requirements:

- Auto-refresh, shown in design as every 5 seconds.
- Manual refresh.
- Tile and list views.
- Search by receipt, product, cashier, terminal, or operation.
- Store-level KPIs:
  - Today's revenue.
  - Money in drawers.
  - Active terminals.
  - Free terminals.
  - Attention-needed terminals.
  - Offline terminals.
  - Receipt count and average check.
- Terminal cards:
  - POS/SCO number.
  - Status.
  - Current cashier/assistant.
  - Current operation.
  - Current receipt amount.
  - Receipt count.
  - Drawer amount.
  - Last heartbeat.
  - Alerts.
- Event stream:
  - Sale created.
  - Selective control started.
  - Return pending.
  - Payment/fiscal error.
  - Terminal offline.
  - Receipt tape ended.

Clicking a terminal should open terminal details in real time.

## Safe And Cash Movement

The safe section tracks all cash movements.

Dashboard requirements:

- Current safe balance.
- Incoming cash today.
- Outgoing cash today.
- Amount prepared for bank collection.
- Denomination breakdown.
- Safe limit status and predicted breach time.
- Operation journal with filters.

Supported operations:

- Issue change fund to cashier.
- Receive cash from cashier.
- Recount safe.
- Bank collection.
- Business expense.
- Export cash book/accounting report where applicable.

Operation details must show:

- Context and operation type.
- Source and destination.
- Responsible users.
- Denomination breakdown.
- Drawer/safe before and after.
- Related POS-generated request.
- Printed documents.
- Video or attachment references if available.
- Signature status.
- Links to correction documents.

Admin can view operations created on POS or senior cashier terminal. Whether admin can edit
or only review posted operations must be permission-controlled. Posted money operations
should be immutable.

## EoD And Cashier Shift Close

EoD section controls operational day closure.

Requirements:

- Show operational day number, open time, elapsed time, and current progress.
- List all cashiers, POS terminals, SCO terminals, and closure state.
- Show pending blocking checks:
  - Open receipts.
  - Unclosed cashier shifts.
  - Unconfirmed cash collections.
  - Safe recount missing.
  - Bank collection pending.
  - B2B documents missing.
  - Returns missing supporting documents.
  - Fiscal errors.
  - Payment reconciliation mismatch.
  - Marking/inventory sync issues.
- Allow saving draft.
- Allow printing acts.
- Allow confirming collection/close when role allows.
- Export closure package.

Cashier close detail must support:

- Denomination recount.
- Critical operation review.
- Missing document upload.
- Reject operation with reason.
- Confirm documents.
- Show discrepancy between POS expected and counted amount.
- Sign and close shift.

## Users And Roles

User management requirements:

- List users with status, role, terminal access, and last activity.
- Search and filter by role, status, store, and permission set.
- Create/edit/deactivate user.
- Assign roles.
- Configure personnel ID.
- Configure PIN policy.
- Bind iButton/staff card.
- View authentication history.
- View user operation journal.

Role model should support:

- Cashier.
- Senior cashier.
- SCO assistant.
- Store manager.
- Admin.
- Auditor/read-only.
- Integration/service account.

Permissions must be granular enough for cash, returns, device management, products,
pricing, templates, and system settings.

## Products And Prices

Catalog section requirements:

- Product list with search/filter.
- Product status and channel availability.
- Price, tax, category, and barcode data.
- Marking requirement.
- Age restriction.
- Weighted goods behavior.
- Images for SCO/product grid.
- Import/export support.
- Price change history.
- Validation before publishing to terminals.

Product changes must be versioned and distributed to store caches. POS/SCO should know
which catalog version was used for every receipt.

## Reconciliation

Reconciliation compares POS data with external systems.

Sources likely include:

- POS receipt ledger.
- Fiscal device/operator.
- Payment acquirer.
- SBP/QR processor.
- Loyalty/bonus provider.
- Gift card provider.
- B2B/invoice system.
- ERP/inventory backend.
- Marking/Chestny ZNAK style service.
- Bank collection/accounting.

Requirements:

- Show totals by source.
- Highlight mismatches.
- Drill down to receipt/payment/document.
- Mark issue as resolved or create correction task.
- Export reconciliation report.
- Track unresolved issues through EoD.

## Integrations

Integration management requirements:

- Integration cards with provider name, status, last sync, queue size, and errors.
- Enable/disable where safe.
- Configure endpoint and credentials through secure secret management.
- Test connection.
- View recent requests/errors.
- Retry failed jobs.
- Set sync schedule and timeout.

Integration examples visible or implied:

- Payment terminal/acquirer.
- SBP/QR payment.
- Fiscal device/operator.
- Marking validation.
- Loyalty.
- Gift cards.
- ERP/catalog.
- Bank/accounting.
- Video recording references.

## System Settings

System settings cover store and terminal behavior:

- Store identity and legal data.
- Operational day rules.
- POS auto-lock timers.
- Cashier login failed-attempt limits and lockout duration.
- Cash drawer limits.
- Safe limit and bank collection schedule.
- Receipt retention.
- Offline mode limits.
- Approval thresholds.
- Random control percentage.
- SCO language defaults.
- Fiscal printer settings.
- Payment timeout settings.
- Integration retry policies.

The first Store Settings implementation manages authentication hardening values per store:
failed-attempt limit, login lockout duration, and POS idle auto-lock duration. Settings are saved
through Store Edge with a manager session and idempotency key; Store Edge enforces permissions and
uses these values for cashier login lockout. Settings must be versioned. Risky changes should
require confirmation and audit.

Store Settings also exposes a manager-only authentication audit view backed by Store Edge auth
attempt records. The view lists terminal-safe login attempt metadata without raw PINs or credential
tokens and allows an authorized manager to reset an actor lockout with an idempotent command.

## POS And SCO Device Management

Device management requirements:

- List POS and SCO terminals.
- Filter by type, state, zone, store, version, and health.
- Register a new POS/SCO.
- Pair device using code or QR.
- Assign store, zone, hardware profile, layout profile, and role.
- View device health:
  - Online/offline.
  - Current user.
  - Current receipt.
  - Printer/tape state.
  - Scanner state.
  - Payment terminal state.
  - Fiscal device state.
  - Scale state.
  - Software version.
- Block/unblock device.
- Send out of service.
- Trigger configuration reload.

The design shows new KSO modal and a pairing/status screen with code and QR.

## Layout Templates

Layout templates configure POS/SCO product buttons and grids.

Central API: `GET/POST /v1/layout-templates`, `GET/PATCH /v1/layout-templates/{templateId}` with `accentPreset`, `accentColor`, `grid`, and `resolvedAccent*` fields. Admin UI: `/central/layout-templates`. See [`ui-components.md`](../development/ui-components.md).

Requirements:

- List templates.
- Create template.
- Preview grid.
- Assign to store, terminal type, or zone.
- Manage categories.
- Manage tile labels, colors, icons/images, and product links.
- Support empty slots.
- Version and publish changes.
- Validate unavailable products before publishing.

## Payment Methods

Payment method management requirements:

- List methods and providers.
- Configure availability by store, terminal type, and channel.
- Configure priority/order.
- Configure limits and fees.
- Configure split payment rules.
- Configure refund rules.
- Configure terminal/acquirer mapping.
- Enable/disable method.
- View method health and last transactions where applicable.

## Gift Cards

Gift card configuration requirements:

- Configure card product/provider.
- Configure activation, redemption, balance check, and refund rules.
- Configure design/colors for display where relevant.
- Track liability/accounting settings.
- View transactions.

## Receipt Templates

Receipt template requirements:

- Manage multiple template types:
  - Sale.
  - Return.
  - Cash in/out order.
  - Correction.
  - B2B/invoice attachment.
  - SCO digital receipt.
- Preview receipt.
- Configure logo, QR, legal footer, fiscal data placement, payment lines, loyalty block,
  and custom text.
- Version and publish.
- Validate required legal fields.

## Color Schemes And Franchise Branding

Central API: `GET/POST /v1/color-schemes`, `GET/PATCH /v1/color-schemes/{schemeId}`. Admin UI: `/central/color-schemes`. See [`ui-components.md`](../development/ui-components.md).

Branding requirements:

- Manage themes for POS/SCO/admin displays.
- Configure logo, accent color, background, button colors, and receipt styling.
- Assign theme to store/franchise.
- Preview theme against POS, SCO, and receipt templates.
- Prevent unreadable contrast before publish.

## Reports

Reports should include:

- Sales by day, terminal, cashier, category, and payment method.
- Returns and cancellations.
- Cash movements.
- Safe balance history.
- Payment reconciliation.
- Marking exceptions.
- SCO control results.
- EoD summaries.
- User activity.

Exact report set should be finalized with finance and operations.

## Audit And Event Log

The admin event log must capture:

- User login/logout.
- Permission changes.
- Product and price changes.
- Terminal registration and status changes.
- Settings changes.
- Integration changes.
- Cash operations.
- EoD actions.
- Reconciliation decisions.
- Document uploads/rejections.

The event log must be immutable for posted operational events. Corrections should be
additional events.
