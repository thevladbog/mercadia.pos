import { useGetLayoutTemplate } from '@mercadia/api-clients-central';
import {
  ApiError,
  createReceiptFiscalDocument,
  createReceiptPayment,
  type CreateReceiptFiscalDocument202Document,
  type CreateReceiptFiscalDocumentBody,
  type CreateReceiptPayment202Payment,
  type CreateReceiptPaymentBody,
  getCurrentOperationalDay,
  type GetReceipt200,
  GetReceipt200Status,
  getReceipt,
  listReceiptFiscalDocuments,
  type ListReceiptFiscalDocuments200DocumentsItem,
  listReceiptPayments,
  type ListReceiptPayments200PaymentsItem,
  openOperationalDay,
  openReceipt,
  openShift,
  recordTerminalHeartbeat,
  scanReceiptLine,
} from '@mercadia/api-clients-store-edge';
import {
  applyTheme,
  Badge,
  Button,
  Card,
  CardHeading,
  Field,
  Input,
  Label,
  LayoutGrid,
  Numpad,
  Select,
  Tabs,
  TabsList,
  TabsTrigger,
  ThemeProvider,
  type AccentPreset,
  type LayoutGridSpec,
} from '@mercadia/ui';
import { QueryClientProvider } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';
import { I18nextProvider, useTranslation } from 'react-i18next';

import { changeAppLocale, i18n, type AppLocale } from '@/i18n/config.js';
import { queryClient } from '@/query-client.js';

const ALL_CATEGORIES = '__all__';
const PAYMENT_METHOD_CARD_MOCK = 'card_mock' satisfies CreateReceiptPaymentBody['method'];
const PAYMENT_METHOD_CASH = 'cash' satisfies CreateReceiptPaymentBody['method'];
const HEARTBEAT_INTERVAL_MS = 30_000;

type PaymentAttempt = CreateReceiptPaymentBody & {
  idempotencyKey: string;
  receiptId: string;
};

type FiscalAttempt = CreateReceiptFiscalDocumentBody & {
  idempotencyKey: string;
  receiptId: string;
};

type ReceiptPayment = CreateReceiptPayment202Payment | ListReceiptPayments200PaymentsItem;
type ReceiptFiscalDocument =
  | CreateReceiptFiscalDocument202Document
  | ListReceiptFiscalDocuments200DocumentsItem;

type StatusMessage = {
  key: string;
  values?: Record<string, number | string>;
};

function envValue(name: string, fallback: string): string {
  return (import.meta.env[name] as string | undefined) ?? fallback;
}

const terminalConfig = {
  storeId: envValue('VITE_POS_STORE_ID', 'store-1'),
  terminalId: envValue('VITE_POS_TERMINAL_ID', 'pos-1'),
  cashierId: envValue('VITE_POS_CASHIER_ID', 'cashier-1'),
  drawerId: envValue('VITE_POS_DRAWER_ID', 'drawer-1'),
  openedById: envValue('VITE_POS_OPENED_BY_ID', 'admin-1'),
  fiscalDeviceId: envValue('VITE_POS_FISCAL_DEVICE_ID', 'fiscal-1'),
  demoBarcode: envValue('VITE_POS_DEMO_BARCODE', '4600000000000'),
  softwareVersion: envValue('VITE_POS_SOFTWARE_VERSION', 'dev'),
  storeTimeZone: envValue(
    'VITE_POS_STORE_TIME_ZONE',
    Intl.DateTimeFormat().resolvedOptions().timeZone,
  ),
};

function createIdempotencyKey(action: string): string {
  return `pos-terminal:${action}:${crypto.randomUUID()}`;
}

function createIdempotencyHeaders(idempotencyKey: string): HeadersInit {
  return { 'Idempotency-Key': idempotencyKey };
}

function formatMinorAmount(amountMinor: number, language: string): string {
  const locale = language === 'en' ? 'en-US' : 'ru-RU';
  return new Intl.NumberFormat(locale, { style: 'currency', currency: 'RUB' }).format(
    amountMinor / 100,
  );
}

function localCalendarDate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function todayBusinessDate(timeZone: string): string {
  try {
    const parts = new Intl.DateTimeFormat('en-US', {
      timeZone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).formatToParts(new Date());
    const getPart = (type: string) => parts.find((part) => part.type === type)?.value;
    const year = getPart('year');
    const month = getPart('month');
    const day = getPart('day');
    if (year && month && day) {
      return `${year}-${month}-${day}`;
    }
  } catch {
    // Fall back to the terminal-local calendar if a configured IANA zone is invalid.
  }
  return localCalendarDate(new Date());
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) {
    return error.problem.detail ?? `${error.problem.code}: ${error.problem.title}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

function resolveTemplateId(): string {
  const params = new URLSearchParams(window.location.search);
  return params.get('templateId') ?? import.meta.env.VITE_LAYOUT_TEMPLATE_ID ?? '';
}

function filterGridByCategory(grid: LayoutGridSpec, categoryId: string | null): LayoutGridSpec {
  if (categoryId === null) {
    return grid;
  }
  return {
    ...grid,
    tiles: grid.tiles.filter((tile) => tile.categoryId === categoryId),
  };
}

function parseAmountToMinor(value: string): number | null {
  const normalized = value.trim().replace(',', '.');
  if (!normalized) {
    return null;
  }
  if (!/^\d+(?:\.\d{0,2})?$/.test(normalized)) {
    return null;
  }
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }
  const [major, fractional = ''] = normalized.split('.');
  return Number(major) * 100 + Number(fractional.padEnd(2, '0'));
}

function formatInputAmount(amountMinor: number): string {
  return (amountMinor / 100).toFixed(2);
}

function isFiscalReceipt(document: ReceiptFiscalDocument): boolean {
  return document.kind === 'receipt';
}

function TerminalShell() {
  const { t, i18n: activeI18n } = useTranslation();
  const templateId = useMemo(() => resolveTemplateId(), []);
  const templateQuery = useGetLayoutTemplate(templateId, {
    query: { enabled: templateId.length > 0 },
  });
  const template = templateQuery.data?.status === 200 ? templateQuery.data.data.template : null;
  const [numpadValue, setNumpadValue] = useState('');
  const [activeCategoryId, setActiveCategoryId] = useState(ALL_CATEGORIES);
  const [barcode, setBarcode] = useState(terminalConfig.demoBarcode);
  const [paymentMethod, setPaymentMethod] =
    useState<CreateReceiptPaymentBody['method']>(PAYMENT_METHOD_CASH);
  const [paymentAmountInput, setPaymentAmountInput] = useState('');
  const [receipt, setReceipt] = useState<GetReceipt200 | null>(null);
  const [payments, setPayments] = useState<ReceiptPayment[]>([]);
  const [paymentAttempt, setPaymentAttempt] = useState<PaymentAttempt | null>(null);
  const [fiscalDocument, setFiscalDocument] = useState<ReceiptFiscalDocument | null>(null);
  const [fiscalAttempt, setFiscalAttempt] = useState<FiscalAttempt | null>(null);
  const [terminalReady, setTerminalReady] = useState(false);
  const [lastHeartbeatAt, setLastHeartbeatAt] = useState<string | null>(null);
  const [busyActionKey, setBusyActionKey] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<StatusMessage>({
    key: 'pos.status.notPrepared',
  });
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!template) {
      return;
    }
    applyTheme({
      surface: template.kind === 'sco' ? 'sco' : 'terminal',
      colorMode: 'light',
      accentPreset: template.resolvedAccentPreset as AccentPreset,
      accent: template.resolvedAccentColor,
    });
  }, [template]);

  useEffect(() => {
    let cancelled = false;
    async function sendHeartbeat(): Promise<void> {
      try {
        const response = await recordTerminalHeartbeat(
          terminalConfig.terminalId,
          {
            kind: 'pos',
            softwareVersion: terminalConfig.softwareVersion,
            storeId: terminalConfig.storeId,
          },
          { headers: createIdempotencyHeaders(createIdempotencyKey('heartbeat')) },
        );
        if (!cancelled && response.status === 202) {
          setLastHeartbeatAt(response.data.terminal.lastSeenAt);
        }
      } catch {
        if (!cancelled) {
          setLastHeartbeatAt(null);
        }
      }
    }
    void sendHeartbeat();
    const intervalId = window.setInterval(() => void sendHeartbeat(), HEARTBEAT_INTERVAL_MS);
    return () => {
      cancelled = true;
      window.clearInterval(intervalId);
    };
  }, []);

  const grid: LayoutGridSpec = useMemo(
    () =>
      template
        ? {
            rows: template.grid.rows ?? 4,
            cols: template.grid.cols ?? 4,
            categories: (template.grid.categories ?? [])
              .filter((category) => category.id)
              .map((category) => ({
                id: category.id!,
                label: category.label ?? '',
              })),
            tiles: (template.grid.tiles ?? []).map((tile) => ({
              label: tile.label ?? '',
              color: tile.color,
              productId: tile.productId,
              empty: tile.empty,
              categoryId: tile.categoryId,
              iconUrl: tile.iconUrl,
            })),
          }
        : {
            rows: 4,
            cols: 4,
            tiles: [{ label: t('pos.demoItem') }, { label: t('pos.returnItem') }],
          },
    [template, t],
  );

  const categories = useMemo(() => grid.categories ?? [], [grid.categories]);
  const resolvedCategoryId = useMemo(() => {
    if (activeCategoryId === ALL_CATEGORIES) {
      return ALL_CATEGORIES;
    }
    return categories.some((category) => category.id === activeCategoryId)
      ? activeCategoryId
      : ALL_CATEGORIES;
  }, [activeCategoryId, categories]);
  const displayGrid = useMemo(
    () =>
      filterGridByCategory(grid, resolvedCategoryId === ALL_CATEGORIES ? null : resolvedCategoryId),
    [grid, resolvedCategoryId],
  );
  const paidMinor = payments.reduce((total, payment) => total + payment.amountMinor, 0);
  const receiptTotalMinor = receipt?.totalMinor ?? 0;
  const remainingMinor = Math.max(receiptTotalMinor - paidMinor, 0);
  const lineCount = receipt?.lines.length ?? 0;
  const canEditReceipt = receipt?.status === GetReceipt200Status.draft;
  const canCapturePayment =
    !!receipt && receiptTotalMinor > 0 && remainingMinor > 0 && !fiscalDocument && !paymentAttempt;
  const isBusy = busyActionKey !== null;
  const displayedStatus = busyActionKey
    ? t('pos.status.running', { action: t(busyActionKey) })
    : t(statusMessage.key, statusMessage.values);
  const formatAmount = (amountMinor: number) => formatMinorAmount(amountMinor, activeI18n.language);

  async function refreshReceiptState(receiptId: string): Promise<GetReceipt200 | null> {
    const response = await getReceipt(receiptId);
    if (response.status === 200) {
      setReceipt(response.data);
      try {
        const paymentsResponse = await listReceiptPayments(receiptId);
        if (paymentsResponse.status === 200) {
          setPayments(paymentsResponse.data.payments);
        } else {
          setPayments([]);
        }
      } catch {
        setPayments([]);
      }
      try {
        const fiscalResponse = await listReceiptFiscalDocuments(receiptId);
        if (fiscalResponse.status === 200) {
          setFiscalDocument(fiscalResponse.data.documents.find(isFiscalReceipt) ?? null);
        } else {
          setFiscalDocument(null);
        }
      } catch {
        setFiscalDocument(null);
      }
      return response.data;
    }
    return null;
  }

  async function runCommand(actionKey: string, command: () => Promise<void>): Promise<void> {
    setBusyActionKey(actionKey);
    setErrorMessage(null);
    try {
      await command();
    } catch (error) {
      setErrorMessage(getErrorMessage(error, t('common.unknownError')));
    } finally {
      setBusyActionKey(null);
    }
  }

  function clearSaleState(): void {
    setReceipt(null);
    setPayments([]);
    setPaymentAttempt(null);
    setPaymentAmountInput('');
    setFiscalDocument(null);
    setFiscalAttempt(null);
    setBarcode(terminalConfig.demoBarcode);
  }

  async function prepareTerminal(): Promise<void> {
    await runCommand('pos.actions.prepareTerminal', async () => {
      let operationalDayId = '';
      try {
        const currentDay = await getCurrentOperationalDay(terminalConfig.storeId);
        if (currentDay.status === 200) {
          operationalDayId = currentDay.data.id;
        }
      } catch (error) {
        if (!(error instanceof ApiError) || error.status !== 404) {
          throw error;
        }
      }

      if (!operationalDayId) {
        const openedDay = await openOperationalDay(
          {
            businessDate: todayBusinessDate(terminalConfig.storeTimeZone),
            openedById: terminalConfig.openedById,
            storeId: terminalConfig.storeId,
          },
          { headers: createIdempotencyHeaders(createIdempotencyKey('open-day')) },
        );
        if (openedDay.status === 202) {
          operationalDayId = openedDay.data.operationalDay.id;
        }
      }

      try {
        await openShift(
          {
            cashierId: terminalConfig.cashierId,
            drawerId: terminalConfig.drawerId,
            openingCashMinor: 0,
            storeId: terminalConfig.storeId,
            terminalId: terminalConfig.terminalId,
          },
          { headers: createIdempotencyHeaders(createIdempotencyKey('open-shift')) },
        );
      } catch (error) {
        if (!(error instanceof ApiError) || error.status !== 409) {
          throw error;
        }
      }

      setTerminalReady(true);
      setStatusMessage({
        key: 'pos.status.ready',
        values: {
          operationalDayId: operationalDayId || t('pos.status.alreadyOpen'),
        },
      });
    });
  }

  async function openNewReceipt(): Promise<void> {
    await runCommand('pos.actions.openReceipt', async () => {
      const response = await openReceipt(
        {
          cashierId: terminalConfig.cashierId,
          channel: 'pos',
          storeId: terminalConfig.storeId,
          terminalId: terminalConfig.terminalId,
        },
        { headers: createIdempotencyHeaders(createIdempotencyKey('open-receipt')) },
      );
      if (response.status === 202) {
        setReceipt(response.data.receipt);
        setPayments([]);
        setPaymentAttempt(null);
        setPaymentAmountInput(formatInputAmount(response.data.receipt.totalMinor));
        setFiscalDocument(null);
        setFiscalAttempt(null);
        setStatusMessage({
          key: 'pos.status.receiptOpened',
          values: { receiptId: response.data.receipt.id },
        });
      } else {
        setErrorMessage(t('pos.errors.openReceiptFailed'));
      }
    });
  }

  async function scanProduct(scannedBarcode = barcode): Promise<boolean> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforeScanning'));
      return false;
    }
    if (!canEditReceipt) {
      setErrorMessage(t('pos.errors.receiptNotEditable'));
      return false;
    }
    let scanned = false;
    await runCommand('pos.actions.scanProduct', async () => {
      const response = await scanReceiptLine(
        receipt.id,
        { barcode: scannedBarcode, quantity: 1 },
        { headers: createIdempotencyHeaders(createIdempotencyKey('scan-product')) },
      );
      if (response.status === 202) {
        const nextRemainingMinor = Math.max(response.data.receipt.totalMinor - paidMinor, 0);
        setReceipt(response.data.receipt);
        setPaymentAmountInput(nextRemainingMinor > 0 ? formatInputAmount(nextRemainingMinor) : '');
        setStatusMessage({ key: 'pos.status.productScanned', values: { barcode: scannedBarcode } });
        scanned = true;
      } else {
        setErrorMessage(t('pos.errors.scanProductFailed'));
      }
    });
    return scanned;
  }

  async function scanTileProduct(tile: LayoutGridSpec['tiles'][number]): Promise<void> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforeScanning'));
      return;
    }
    if (!canEditReceipt) {
      setErrorMessage(t('pos.errors.receiptNotEditable'));
      return;
    }
    if (await scanProduct()) {
      setStatusMessage({ key: 'pos.status.tileScanned', values: { product: tile.label, barcode } });
    }
  }

  async function capturePayment(): Promise<void> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforePayment'));
      return;
    }
    if (receiptTotalMinor <= 0) {
      setErrorMessage(t('pos.errors.positiveReceiptTotalRequired'));
      return;
    }
    const requestedAmount = parseAmountToMinor(paymentAmountInput);
    if (!requestedAmount) {
      setErrorMessage(t('pos.errors.invalidPaymentAmount'));
      return;
    }
    if (requestedAmount > remainingMinor) {
      setErrorMessage(t('pos.errors.paymentAmountTooHigh'));
      return;
    }
    const attempt: PaymentAttempt = {
      amountMinor: requestedAmount,
      idempotencyKey: createIdempotencyKey('capture-payment'),
      method: paymentMethod,
      providerReference:
        paymentMethod === PAYMENT_METHOD_CARD_MOCK ? `POS-${Date.now()}` : undefined,
      receiptId: receipt.id,
    };
    setPaymentAttempt(attempt);
    await runCommand('pos.actions.capturePayment', async () => {
      let response: Awaited<ReturnType<typeof createReceiptPayment>>;
      try {
        response = await createReceiptPayment(
          attempt.receiptId,
          {
            amountMinor: attempt.amountMinor,
            method: attempt.method,
            providerReference: attempt.providerReference,
          },
          { headers: createIdempotencyHeaders(attempt.idempotencyKey) },
        );
      } catch (error) {
        setPaymentAttempt(null);
        throw error;
      }
      if (response.status === 202) {
        setPaymentAttempt(null);
        const refreshedReceipt = await refreshReceiptState(attempt.receiptId);
        const nextRemainingMinor = refreshedReceipt
          ? Math.max(refreshedReceipt.totalMinor - paidMinor - attempt.amountMinor, 0)
          : Math.max(remainingMinor - attempt.amountMinor, 0);
        setPaymentAmountInput(nextRemainingMinor > 0 ? formatInputAmount(nextRemainingMinor) : '');
        setStatusMessage({
          key: 'pos.status.paymentCaptured',
          values: {
            amount: formatAmount(attempt.amountMinor),
            method:
              attempt.method === PAYMENT_METHOD_CARD_MOCK
                ? t('pos.paymentMethods.cardMock')
                : t('pos.paymentMethods.cash'),
          },
        });
      } else {
        setPaymentAttempt(null);
        setErrorMessage(t('pos.errors.capturePaymentFailed'));
      }
    });
  }

  async function fiscalizeReceipt(): Promise<void> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforeFiscalization'));
      return;
    }
    if (receipt.status !== GetReceipt200Status.paid) {
      setErrorMessage(t('pos.errors.paidReceiptRequired'));
      return;
    }
    if (fiscalDocument) {
      setErrorMessage(t('pos.errors.fiscalAlreadySubmitted'));
      return;
    }
    const attempt: FiscalAttempt =
      fiscalAttempt?.receiptId === receipt.id
        ? fiscalAttempt
        : {
            deviceId: terminalConfig.fiscalDeviceId,
            idempotencyKey: createIdempotencyKey('fiscalize-receipt'),
            receiptId: receipt.id,
          };
    setFiscalAttempt(attempt);
    await runCommand('pos.actions.fiscalize', async () => {
      const response = await createReceiptFiscalDocument(
        attempt.receiptId,
        { deviceId: attempt.deviceId },
        { headers: createIdempotencyHeaders(attempt.idempotencyKey) },
      );
      if (response.status === 202) {
        setFiscalDocument(response.data.document);
        setFiscalAttempt(null);
        await refreshReceiptState(attempt.receiptId);
        setStatusMessage({
          key: 'pos.status.fiscalCreated',
          values: { documentId: response.data.document.id },
        });
      } else {
        setErrorMessage(t('pos.errors.fiscalizationFailed'));
      }
    });
  }

  function finishSale(): void {
    clearSaleState();
    setStatusMessage({ key: 'pos.status.saleFinished' });
    setErrorMessage(null);
  }

  function handleTileClick(_index: number, tile: LayoutGridSpec['tiles'][number]): void {
    void scanTileProduct(tile);
  }

  return (
    <main className="pos-cockpit-shell">
      <header className="pos-cockpit-header">
        <div>
          <p className="eyebrow">{t('pos.eyebrow')}</p>
          <h1>{template?.name ?? t('pos.titleFallback')}</h1>
          <p className="muted">
            {template
              ? `${template.kind} · ${template.resolvedAccentColor}`
              : t('pos.templateHint')}
          </p>
        </div>
        <div className="pos-header-actions">
          <Field className="language-select">
            <Label htmlFor="pos-language-select">{t('language.label')}</Label>
            <Select
              id="pos-language-select"
              value={activeI18n.language}
              onChange={(event) => changeAppLocale(event.target.value as AppLocale)}
            >
              <option value="ru">{t('language.ru')}</option>
              <option value="en">{t('language.en')}</option>
            </Select>
          </Field>
          <Button type="button" disabled={isBusy} onClick={() => void prepareTerminal()}>
            {terminalReady ? t('pos.actions.refreshReadiness') : t('pos.actions.prepareTerminal')}
          </Button>
        </div>
      </header>

      <section className="terminal-strip" aria-label={t('pos.terminalState')}>
        <div>
          <span>{t('pos.store')}</span>
          <strong>{terminalConfig.storeId}</strong>
        </div>
        <div>
          <span>{t('pos.terminal')}</span>
          <strong>{terminalConfig.terminalId}</strong>
        </div>
        <div>
          <span>{t('pos.cashier')}</span>
          <strong>{terminalConfig.cashierId}</strong>
        </div>
        <div>
          <span>{t('pos.drawer')}</span>
          <strong>{terminalConfig.drawerId}</strong>
        </div>
        <div>
          <span>{t('pos.heartbeat')}</span>
          <strong>{lastHeartbeatAt ? t('pos.heartbeatOnline') : t('pos.heartbeatPending')}</strong>
        </div>
      </section>

      <section
        aria-atomic="true"
        aria-live={errorMessage ? 'assertive' : 'polite'}
        className="status-line"
        data-ready={terminalReady}
        role={errorMessage ? 'alert' : 'status'}
      >
        <span>{displayedStatus}</span>
        {errorMessage ? <strong>{errorMessage}</strong> : null}
      </section>

      <section className="cockpit-grid">
        <Card className="receipt-card">
          <CardHeading
            title={receipt ? t('pos.receipt', { id: receipt.id }) : t('pos.noActiveReceipt')}
            subtitle={receipt ? t('pos.receiptDescription') : t('pos.openReceiptHint')}
          />
          <div className="receipt-toolbar">
            <Button
              type="button"
              disabled={isBusy || !terminalReady || receipt !== null}
              onClick={() => void openNewReceipt()}
            >
              {t('pos.actions.startSale')}
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={isBusy || !receipt || !fiscalDocument}
              onClick={finishSale}
            >
              {t('pos.actions.finishSale')}
            </Button>
            <Badge variant={receipt ? 'accent' : 'outline'}>
              {receipt?.status ?? t('pos.status.noReceipt')}
            </Badge>
            <Badge variant={lineCount > 0 ? 'success' : 'outline'}>
              {t('pos.linesCount', { count: lineCount })}
            </Badge>
          </div>
          <div className="receipt-table-wrap">
            <table className="receipt-table">
              <thead>
                <tr>
                  <th>{t('pos.product')}</th>
                  <th>{t('pos.quantity')}</th>
                  <th>{t('pos.price')}</th>
                  <th>{t('pos.total')}</th>
                </tr>
              </thead>
              <tbody>
                {!receipt || receipt.lines.length === 0 ? (
                  <tr>
                    <td colSpan={4}>{t('pos.noLines')}</td>
                  </tr>
                ) : (
                  receipt.lines.map((line) => (
                    <tr key={line.id}>
                      <td>
                        <strong>{line.name}</strong>
                        <span>{line.barcode || line.productId}</span>
                      </td>
                      <td>{line.quantity}</td>
                      <td>{formatAmount(line.unitPriceMinor)}</td>
                      <td>{formatAmount(line.totalMinor)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>

        <aside className="totals-rail">
          <Card className="total-card total-card--primary">
            <span>{t('pos.totalDue')}</span>
            <strong>{formatAmount(receiptTotalMinor)}</strong>
          </Card>
          <Card className="total-card">
            <span>{t('pos.paid')}</span>
            <strong>{formatAmount(paidMinor)}</strong>
          </Card>
          <Card className="total-card">
            <span>{t('pos.remaining')}</span>
            <strong>{formatAmount(remainingMinor)}</strong>
          </Card>
          <Card className="payment-card">
            <CardHeading title={t('pos.payment')} subtitle={t('pos.paymentDescription')} />
            <Field>
              <Label htmlFor="pos-payment-method-select">{t('pos.paymentMethod')}</Label>
              <Select
                id="pos-payment-method-select"
                value={paymentMethod}
                disabled={isBusy || !canCapturePayment}
                onChange={(event) =>
                  setPaymentMethod(event.target.value as CreateReceiptPaymentBody['method'])
                }
              >
                <option value={PAYMENT_METHOD_CASH}>{t('pos.paymentMethods.cash')}</option>
                <option value={PAYMENT_METHOD_CARD_MOCK}>{t('pos.paymentMethods.cardMock')}</option>
              </Select>
            </Field>
            <Field>
              <Label htmlFor="pos-payment-amount-input">{t('pos.paymentAmount')}</Label>
              <Input
                id="pos-payment-amount-input"
                inputMode="decimal"
                value={paymentAmountInput}
                disabled={isBusy || !canCapturePayment}
                onChange={(event) => setPaymentAmountInput(event.target.value)}
              />
            </Field>
            <Button
              type="button"
              disabled={isBusy || !canCapturePayment}
              onClick={() => void capturePayment()}
            >
              {t('pos.actions.capturePayment')}
            </Button>
            <div className="payment-list">
              {payments.length === 0 ? (
                <span className="muted">{t('pos.noPayments')}</span>
              ) : (
                payments.map((payment) => (
                  <span key={payment.id}>
                    {payment.method === PAYMENT_METHOD_CARD_MOCK
                      ? t('pos.paymentMethods.cardMock')
                      : t('pos.paymentMethods.cash')}{' '}
                    · {formatAmount(payment.amountMinor)} · {payment.status}
                  </span>
                ))
              )}
            </div>
          </Card>
          <Card className="payment-card">
            <CardHeading title={t('pos.fiscal')} subtitle={t('pos.fiscalDescription')} />
            <Button
              type="button"
              disabled={
                isBusy ||
                !receipt ||
                receipt.status !== GetReceipt200Status.paid ||
                fiscalDocument !== null
              }
              onClick={() => void fiscalizeReceipt()}
            >
              {t('pos.actions.fiscalize')}
            </Button>
            <span className="muted">
              {fiscalDocument
                ? `${fiscalDocument.status} · ${fiscalDocument.fiscalSign}`
                : t('pos.noFiscalDocument')}
            </span>
          </Card>
        </aside>
      </section>

      <section className="workbench-grid">
        <Card className="scanner-card">
          <CardHeading title={t('pos.scanner')} subtitle={t('pos.scannerDescription')} />
          <div className="scanner-row">
            <Field>
              <Label htmlFor="pos-barcode-input">{t('pos.barcode')}</Label>
              <Input
                id="pos-barcode-input"
                value={barcode}
                disabled={isBusy || !canEditReceipt}
                onChange={(event) => setBarcode(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    void scanProduct();
                  }
                }}
              />
            </Field>
            <Button
              type="button"
              disabled={isBusy || !receipt || !canEditReceipt}
              onClick={() => void scanProduct()}
            >
              {t('pos.actions.scanProduct')}
            </Button>
          </div>
          {!templateId ? (
            <p className="muted">{t('pos.templateMissing')}</p>
          ) : templateQuery.isLoading ? (
            <p className="muted">{t('pos.templateLoading')}</p>
          ) : templateQuery.isError ? (
            <p className="error">{t('pos.templateLoadError')}</p>
          ) : null}
          {categories.length > 0 ? (
            <Tabs value={resolvedCategoryId} onValueChange={setActiveCategoryId}>
              <TabsList aria-label={t('pos.categories')}>
                <TabsTrigger value={ALL_CATEGORIES}>{t('pos.allCategories')}</TabsTrigger>
                {categories.map((category) => (
                  <TabsTrigger key={category.id} value={category.id}>
                    {category.label}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          ) : null}
          <LayoutGrid grid={displayGrid} onTileClick={handleTileClick} />
        </Card>
        <Card className="numpad-card">
          <CardHeading title={t('pos.numpad')} subtitle={t('pos.numpadDescription')} />
          <Numpad enterLabel={t('pos.numpadEnter')} value={numpadValue} onChange={setNumpadValue} />
        </Card>
      </section>
    </main>
  );
}

export function Root() {
  return (
    <ThemeProvider defaultTheme={{ surface: 'terminal', colorMode: 'light', accentPreset: 'sale' }}>
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <TerminalShell />
        </QueryClientProvider>
      </I18nextProvider>
    </ThemeProvider>
  );
}
