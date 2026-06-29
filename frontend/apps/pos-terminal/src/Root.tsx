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
  openOperationalDay,
  openReceipt,
  openShift,
  scanReceiptLine,
} from '@mercadia/api-clients-store-edge';
import {
  applyTheme,
  Button,
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

type PaymentAttempt = CreateReceiptPaymentBody & {
  idempotencyKey: string;
  receiptId: string;
};

type FiscalAttempt = CreateReceiptFiscalDocumentBody & {
  idempotencyKey: string;
  receiptId: string;
};

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
  const [receipt, setReceipt] = useState<GetReceipt200 | null>(null);
  const [payment, setPayment] = useState<CreateReceiptPayment202Payment | null>(null);
  const [paymentAttempt, setPaymentAttempt] = useState<PaymentAttempt | null>(null);
  const [fiscalDocument, setFiscalDocument] =
    useState<CreateReceiptFiscalDocument202Document | null>(null);
  const [fiscalAttempt, setFiscalAttempt] = useState<FiscalAttempt | null>(null);
  const [terminalReady, setTerminalReady] = useState(false);
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

  async function refreshReceipt(receiptId: string): Promise<void> {
    const response = await getReceipt(receiptId);
    if (response.status === 200) {
      setReceipt(response.data);
    }
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
        setPayment(null);
        setPaymentAttempt(null);
        setFiscalDocument(null);
        setFiscalAttempt(null);
        setStatusMessage({
          key: 'pos.status.receiptOpened',
          values: { receiptId: response.data.receipt.id },
        });
      }
    });
  }

  async function scanDemoProduct(): Promise<void> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforeScanning'));
      return;
    }
    if (payment || paymentAttempt) {
      setErrorMessage(t('pos.errors.paymentAlreadySubmitted'));
      return;
    }
    await runCommand('pos.actions.scanProduct', async () => {
      const response = await scanReceiptLine(
        receipt.id,
        { barcode, quantity: 1 },
        { headers: createIdempotencyHeaders(createIdempotencyKey('scan-product')) },
      );
      if (response.status === 202) {
        setReceipt(response.data.receipt);
        setStatusMessage({ key: 'pos.status.productScanned', values: { barcode } });
      }
    });
  }

  async function capturePayment(): Promise<void> {
    if (!receipt) {
      setErrorMessage(t('pos.errors.openReceiptBeforePayment'));
      return;
    }
    if (receipt.totalMinor <= 0) {
      setErrorMessage(t('pos.errors.positiveReceiptTotalRequired'));
      return;
    }
    if (payment) {
      setErrorMessage(t('pos.errors.paymentAlreadySubmitted'));
      return;
    }
    const attempt: PaymentAttempt =
      paymentAttempt?.receiptId === receipt.id
        ? paymentAttempt
        : {
            amountMinor: receipt.totalMinor,
            idempotencyKey: createIdempotencyKey('capture-payment'),
            method: paymentMethod,
            providerReference:
              paymentMethod === PAYMENT_METHOD_CARD_MOCK ? `POS-${Date.now()}` : undefined,
            receiptId: receipt.id,
          };
    setPaymentAttempt(attempt);
    await runCommand('pos.actions.capturePayment', async () => {
      const response = await createReceiptPayment(
        attempt.receiptId,
        {
          amountMinor: attempt.amountMinor,
          method: attempt.method,
          providerReference: attempt.providerReference,
        },
        { headers: createIdempotencyHeaders(attempt.idempotencyKey) },
      );
      if (response.status === 202) {
        setPayment(response.data.payment);
        setPaymentAttempt(null);
        await refreshReceipt(attempt.receiptId);
        setStatusMessage({
          key: 'pos.status.paymentCaptured',
          values: {
            method:
              attempt.method === PAYMENT_METHOD_CARD_MOCK
                ? t('pos.paymentMethods.cardMock')
                : t('pos.paymentMethods.cash'),
          },
        });
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
        await refreshReceipt(attempt.receiptId);
        setStatusMessage({
          key: 'pos.status.fiscalCreated',
          values: { documentId: response.data.document.id },
        });
      }
    });
  }

  const isBusy = busyActionKey !== null;
  const displayedStatus = busyActionKey
    ? t('pos.status.running', { action: t(busyActionKey) })
    : t(statusMessage.key, statusMessage.values);
  const formatAmount = (amountMinor: number) => formatMinorAmount(amountMinor, activeI18n.language);
  const paymentDisplay = payment
    ? `${
        payment.method === PAYMENT_METHOD_CARD_MOCK
          ? t('pos.paymentMethods.cardMock')
          : t('pos.paymentMethods.cash')
      } · ${payment.status}`
    : t('common.none');

  return (
    <main className="pos-terminal-shell">
      <header className="pos-terminal-header">
        <div className="pos-terminal-title-row">
          <div>
            <h1>{template?.name ?? t('pos.titleFallback')}</h1>
            <p className="muted">
              {template
                ? `${template.kind} · ${template.resolvedAccentColor}`
                : t('pos.templateHint')}
            </p>
          </div>
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
        </div>
        <dl className="terminal-meta">
          <div>
            <dt>{t('pos.store')}</dt>
            <dd>{terminalConfig.storeId}</dd>
          </div>
          <div>
            <dt>{t('pos.terminal')}</dt>
            <dd>{terminalConfig.terminalId}</dd>
          </div>
          <div>
            <dt>{t('pos.cashier')}</dt>
            <dd>{terminalConfig.cashierId}</dd>
          </div>
        </dl>
      </header>
      <section className="checkout-panel panel">
        <div>
          <h2>{t('pos.checkoutTitle')}</h2>
          <p className="muted">{t('pos.checkoutDescription')}</p>
        </div>
        <div className="checkout-actions">
          <Button type="button" disabled={isBusy} onClick={() => void prepareTerminal()}>
            {t('pos.actions.prepareTerminal')}
          </Button>
          <Button type="button" disabled={isBusy} onClick={() => void openNewReceipt()}>
            {t('pos.actions.openReceipt')}
          </Button>
          <Button
            type="button"
            disabled={isBusy || !receipt || payment !== null || paymentAttempt !== null}
            onClick={() => void scanDemoProduct()}
          >
            {t('pos.actions.scanProduct')}
          </Button>
          <Button
            type="button"
            disabled={isBusy || !receipt || receipt.totalMinor <= 0 || payment !== null}
            onClick={() => void capturePayment()}
          >
            {t('pos.actions.capturePayment')}
          </Button>
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
        </div>
        <div className="checkout-form-row">
          <Field className="checkout-field">
            <Label htmlFor="pos-barcode-input">{t('pos.barcode')}</Label>
            <Input
              id="pos-barcode-input"
              value={barcode}
              onChange={(event) => setBarcode(event.target.value)}
            />
          </Field>
          <Field className="checkout-field">
            <Label htmlFor="pos-payment-method-select">{t('pos.paymentMethod')}</Label>
            <Select
              id="pos-payment-method-select"
              value={paymentMethod}
              disabled={isBusy || payment !== null || paymentAttempt !== null}
              onChange={(event) => setPaymentMethod(event.target.value)}
            >
              <option value={PAYMENT_METHOD_CASH}>{t('pos.paymentMethods.cash')}</option>
              <option value={PAYMENT_METHOD_CARD_MOCK}>{t('pos.paymentMethods.cardMock')}</option>
            </Select>
          </Field>
        </div>
        <div className="status-line" data-ready={terminalReady}>
          {displayedStatus}
        </div>
        {errorMessage ? <div className="error-box">{errorMessage}</div> : null}
        {receipt ? (
          <div className="receipt-board">
            <div className="receipt-summary">
              <strong>{t('pos.receipt', { id: receipt.id })}</strong>
              <span>{receipt.status}</span>
              <span>{formatAmount(receipt.totalMinor)}</span>
            </div>
            <table>
              <thead>
                <tr>
                  <th>{t('pos.product')}</th>
                  <th>{t('pos.quantity')}</th>
                  <th>{t('pos.price')}</th>
                  <th>{t('pos.total')}</th>
                </tr>
              </thead>
              <tbody>
                {receipt.lines.length === 0 ? (
                  <tr>
                    <td colSpan={4}>{t('pos.noLines')}</td>
                  </tr>
                ) : (
                  receipt.lines.map((line) => (
                    <tr key={line.id}>
                      <td>{line.name}</td>
                      <td>{line.quantity}</td>
                      <td>{formatAmount(line.unitPriceMinor)}</td>
                      <td>{formatAmount(line.totalMinor)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
            <div className="receipt-events">
              <span>
                {t('pos.payment')}: {paymentDisplay}
              </span>
              <span>
                {t('pos.fiscal')}:{' '}
                {fiscalDocument
                  ? `${fiscalDocument.status} · ${fiscalDocument.fiscalSign}`
                  : t('common.none')}
              </span>
            </div>
          </div>
        ) : null}
      </section>
      {!templateId ? (
        <p className="muted">{t('pos.templateMissing')}</p>
      ) : templateQuery.isLoading ? (
        <p className="muted">{t('pos.templateLoading')}</p>
      ) : templateQuery.isError ? (
        <p className="error">{t('pos.templateLoadError')}</p>
      ) : (
        <div className="pos-terminal-grid">
          <section className="panel">
            <Button type="button" disabled={isBusy} onClick={() => void openNewReceipt()}>
              {t('pos.actions.startSale')}
            </Button>
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
            <LayoutGrid grid={displayGrid} onTileClick={() => void scanDemoProduct()} />
          </section>
          <section className="panel">
            <Numpad
              enterLabel={t('pos.numpadEnter')}
              value={numpadValue}
              onChange={setNumpadValue}
            />
          </section>
        </div>
      )}
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
