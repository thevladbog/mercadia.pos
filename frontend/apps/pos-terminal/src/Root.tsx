import { useGetLayoutTemplate } from '@mercadia/api-clients-central';
import {
  ApiError,
  createReceiptFiscalDocument,
  createReceiptPayment,
  getCurrentOperationalDay,
  getReceipt,
  openOperationalDay,
  openReceipt,
  openShift,
  scanReceiptLine,
} from '@mercadia/api-clients-store-edge';
import {
  applyTheme,
  Button,
  LayoutGrid,
  Numpad,
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

type ReceiptLine = {
  id: string;
  name: string;
  productId: string;
  quantity: number;
  totalMinor: number;
  unitPriceMinor: number;
};

type ReceiptView = {
  id: string;
  lines: ReceiptLine[];
  status: string;
  totalMinor: number;
};

type PaymentView = {
  amountMinor: number;
  id: string;
  method: string;
  status: string;
};

type FiscalDocumentView = {
  fiscalSign: string;
  id: string;
  status: string;
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
};

function createIdempotencyHeaders(action: string): HeadersInit {
  return { 'Idempotency-Key': `pos-terminal:${action}:${crypto.randomUUID()}` };
}

function formatMinorAmount(amountMinor: number, language: string): string {
  const locale = language === 'en' ? 'en-US' : 'ru-RU';
  return new Intl.NumberFormat(locale, { style: 'currency', currency: 'RUB' }).format(
    amountMinor / 100,
  );
}

function todayBusinessDate(): string {
  return new Date().toISOString().slice(0, 10);
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
  const [paymentMethod, setPaymentMethod] = useState('cash');
  const [receipt, setReceipt] = useState<ReceiptView | null>(null);
  const [payment, setPayment] = useState<PaymentView | null>(null);
  const [fiscalDocument, setFiscalDocument] = useState<FiscalDocumentView | null>(null);
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
            businessDate: todayBusinessDate(),
            openedById: terminalConfig.openedById,
            storeId: terminalConfig.storeId,
          },
          { headers: createIdempotencyHeaders('open-day') },
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
          { headers: createIdempotencyHeaders('open-shift') },
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
        { headers: createIdempotencyHeaders('open-receipt') },
      );
      if (response.status === 202) {
        setReceipt(response.data.receipt);
        setPayment(null);
        setFiscalDocument(null);
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
    await runCommand('pos.actions.scanProduct', async () => {
      const response = await scanReceiptLine(
        receipt.id,
        { barcode, quantity: 1 },
        { headers: createIdempotencyHeaders('scan-product') },
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
    await runCommand('pos.actions.capturePayment', async () => {
      const response = await createReceiptPayment(
        receipt.id,
        {
          amountMinor: receipt.totalMinor,
          method: paymentMethod,
          providerReference: paymentMethod === 'card_mock' ? `POS-${Date.now()}` : undefined,
        },
        { headers: createIdempotencyHeaders('capture-payment') },
      );
      if (response.status === 202) {
        setPayment(response.data.payment);
        await refreshReceipt(receipt.id);
        setStatusMessage({
          key: 'pos.status.paymentCaptured',
          values: {
            method:
              paymentMethod === 'card_mock'
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
    await runCommand('pos.actions.fiscalize', async () => {
      const response = await createReceiptFiscalDocument(
        receipt.id,
        { deviceId: terminalConfig.fiscalDeviceId },
        { headers: createIdempotencyHeaders('fiscalize-receipt') },
      );
      if (response.status === 202) {
        setFiscalDocument(response.data.document);
        await refreshReceipt(receipt.id);
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
        payment.method === 'card_mock'
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
          <label className="language-select">
            {t('language.label')}
            <select
              value={activeI18n.language}
              onChange={(event) => changeAppLocale(event.target.value as AppLocale)}
            >
              <option value="ru">{t('language.ru')}</option>
              <option value="en">{t('language.en')}</option>
            </select>
          </label>
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
            disabled={isBusy || !receipt}
            onClick={() => void scanDemoProduct()}
          >
            {t('pos.actions.scanProduct')}
          </Button>
          <Button
            type="button"
            disabled={isBusy || !receipt || receipt.totalMinor <= 0}
            onClick={() => void capturePayment()}
          >
            {t('pos.actions.capturePayment')}
          </Button>
          <Button
            type="button"
            disabled={isBusy || !receipt || receipt.status !== 'paid'}
            onClick={() => void fiscalizeReceipt()}
          >
            {t('pos.actions.fiscalize')}
          </Button>
        </div>
        <div className="checkout-form-row">
          <label>
            {t('pos.barcode')}
            <input value={barcode} onChange={(event) => setBarcode(event.target.value)} />
          </label>
          <label>
            {t('pos.paymentMethod')}
            <select
              value={paymentMethod}
              onChange={(event) => setPaymentMethod(event.target.value)}
            >
              <option value="cash">{t('pos.paymentMethods.cash')}</option>
              <option value="card_mock">{t('pos.paymentMethods.cardMock')}</option>
            </select>
          </label>
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
            <Button type="button">{t('pos.actions.startSale')}</Button>
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
