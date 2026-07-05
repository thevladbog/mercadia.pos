import { useEffect, useState } from 'react';
import {
  ApiError,
  cancelReceipt,
  getCurrentOperationalDay,
  type GetReceipt200,
  GetReceipt200Status,
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
  DetailDialog,
  Field,
  Input,
  Label,
  Select,
  ThemeProvider,
} from '@mercadia/ui';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider, useTranslation } from 'react-i18next';

import {
  createIdempotencyHeaders,
  createIdempotencyKey,
  formatMinorAmount,
} from '@mercadia/receipt-kit';

import { AuthProvider, useAuth } from '@/auth/AuthProvider.js';
import type { SessionResult } from '@/auth/types.js';
import { envValue, getStoreId, getTerminalId } from '@/api-client-config.js';
import { changeAppLocale, i18n, type AppLocale } from '@/i18n/config.js';
import { parseLayoutProfile } from '@/layout-profile.js';
import { queryClient } from '@/query-client.js';
import {
  isStageComingSoon,
  reduceScoStage,
  STAGE_INDICATOR_ORDER,
  type ScoStage,
} from '@/stage.js';

const HEARTBEAT_INTERVAL_MS = 30_000;

const terminalConfig = {
  storeId: getStoreId(),
  terminalId: getTerminalId(),
  // Store-Edge's OpenReceipt requires an open shift for the terminal (checkout.go), which in
  // turn requires a drawerId even though SCO is cashless (ADR-0008): the ledger path is only
  // exercised when openingCashMinor > 0, so a placeholder drawer with a 0 opening balance has no
  // cash-reconciliation side effects.
  drawerId: envValue('VITE_SCO_DRAWER_ID', 'sco-drawer-1'),
  demoBarcode: envValue('VITE_SCO_DEMO_BARCODE', '4600000000000'),
  softwareVersion: envValue('VITE_SCO_SOFTWARE_VERSION', 'dev'),
  layoutProfile: parseLayoutProfile(import.meta.env.VITE_SCO_LAYOUT_PROFILE),
};

function todayBusinessDate(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

type PrepStatus = 'preparing' | 'ready' | 'error';

function BootScreen({ messageKey }: { messageKey: string }) {
  const { t } = useTranslation();
  return (
    <main className="sco-boot-shell" role="status" aria-live="polite">
      <p className="muted">{t(messageKey)}</p>
    </main>
  );
}

function OutOfServiceScreen({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslation();
  return (
    <main className="sco-boot-shell" role="alert">
      <Card className="sco-out-of-service-card">
        <CardHeading title={t('sco.outOfService.title')} subtitle={t('sco.outOfService.message')} />
        <Button type="button" onClick={onRetry}>
          {t('sco.outOfService.retry')}
        </Button>
      </Card>
    </main>
  );
}

function LanguageSelect() {
  const { t, i18n: activeI18n } = useTranslation();
  return (
    <Field className="sco-language-select">
      <Label htmlFor="sco-language-select">{t('language.label')}</Label>
      <Select
        id="sco-language-select"
        value={activeI18n.language}
        onChange={(event) => changeAppLocale(event.target.value as AppLocale)}
      >
        <option value="ru">{t('language.ru')}</option>
        <option value="en">{t('language.en')}</option>
      </Select>
    </Field>
  );
}

function Terminal({ session }: { session: SessionResult }) {
  const { t, i18n: activeI18n } = useTranslation();
  const [prepStatus, setPrepStatus] = useState<PrepStatus>('preparing');
  const [stage, setStage] = useState<ScoStage>('idle');
  const [receipt, setReceipt] = useState<GetReceipt200 | null>(null);
  const [barcode, setBarcode] = useState(terminalConfig.demoBarcode);
  const [isBusy, setIsBusy] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [lastHeartbeatAt, setLastHeartbeatAt] = useState<string | null>(null);
  const [prepAttempt, setPrepAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    async function prepareTerminal(): Promise<void> {
      try {
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
          await openOperationalDay(
            {
              businessDate: todayBusinessDate(),
              openedById: session.actorId,
              storeId: terminalConfig.storeId,
            },
            { headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'open-day')) },
          );
        }

        try {
          await openShift(
            {
              cashierId: session.actorId,
              drawerId: terminalConfig.drawerId,
              openingCashMinor: 0,
              storeId: terminalConfig.storeId,
              terminalId: terminalConfig.terminalId,
            },
            {
              headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'open-shift')),
            },
          );
        } catch (error) {
          if (!(error instanceof ApiError) || error.status !== 409) {
            throw error;
          }
        }

        if (!cancelled) {
          setPrepStatus('ready');
        }
      } catch (error) {
        console.error('[sco-terminal] terminal preparation failed', error);
        if (!cancelled) {
          setPrepStatus('error');
        }
      }
    }
    void prepareTerminal();
    return () => {
      cancelled = true;
    };
  }, [session.actorId, prepAttempt]);

  function retryPreparation(): void {
    setPrepStatus('preparing');
    setPrepAttempt((attempt) => attempt + 1);
  }

  useEffect(() => {
    applyTheme({
      surface: 'sco',
      colorMode: 'light',
      accentPreset: 'sco',
    });
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function sendHeartbeat(): Promise<void> {
      try {
        const response = await recordTerminalHeartbeat(
          terminalConfig.terminalId,
          {
            kind: 'sco',
            softwareVersion: terminalConfig.softwareVersion,
            storeId: terminalConfig.storeId,
          },
          {
            headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'heartbeat')),
          },
        );
        if (!cancelled && response.status === 202) {
          setLastHeartbeatAt(response.data.terminal.lastSeenAt);
        }
      } catch (error) {
        console.error('[sco-terminal] heartbeat failed', error);
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

  // Defensive guard: the idle screen must never show or act on a leftover cart, even if a
  // receipt somehow survives a stage transition back to idle. Deriving it at render time (rather
  // than reactively clearing it in an effect) means there is no path — however state gets there —
  // that lets stale receipt data reach the idle screen.
  const activeReceipt = stage === 'idle' ? null : receipt;

  const lineCount = activeReceipt?.lines.length ?? 0;
  const totalMinor = activeReceipt?.totalMinor ?? 0;
  const formatAmount = (amountMinor: number) => formatMinorAmount(amountMinor, activeI18n.language);
  const canEditReceipt = !activeReceipt || activeReceipt.status === GetReceipt200Status.draft;

  async function startCheckout(): Promise<void> {
    if (isBusy) return;
    setErrorMessage(null);
    setIsBusy(true);
    try {
      const response = await openReceipt(
        {
          cashierId: session.actorId,
          channel: 'sco',
          storeId: terminalConfig.storeId,
          terminalId: terminalConfig.terminalId,
        },
        {
          headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'open-receipt')),
        },
      );
      if (response.status === 202) {
        setReceipt(response.data.receipt);
        setStage((prev) => reduceScoStage(prev, { type: 'start' }));
      } else {
        setErrorMessage(t('sco.errors.startFailed'));
      }
    } catch (error) {
      console.error('[sco-terminal] open receipt failed', error);
      setErrorMessage(t('sco.errors.startFailed'));
    } finally {
      setIsBusy(false);
    }
  }

  async function scanBarcode(): Promise<void> {
    if (isBusy || !receipt || !canEditReceipt || !barcode.trim()) return;
    setErrorMessage(null);
    setIsBusy(true);
    try {
      const response = await scanReceiptLine(
        receipt.id,
        { barcode: barcode.trim(), quantity: 1 },
        {
          headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'scan-line')),
        },
      );
      if (response.status === 202) {
        setReceipt(response.data.receipt);
      } else {
        setErrorMessage(t('sco.errors.scanFailed'));
      }
    } catch (error) {
      console.error('[sco-terminal] scan line failed', error);
      setErrorMessage(t('sco.errors.scanFailed'));
    } finally {
      setIsBusy(false);
    }
  }

  function scanDemoBarcode(): void {
    setBarcode(terminalConfig.demoBarcode);
    void scanBarcode();
  }

  async function confirmCancel(): Promise<void> {
    if (!receipt) {
      setCancelDialogOpen(false);
      return;
    }
    setErrorMessage(null);
    setIsBusy(true);
    try {
      await cancelReceipt(
        receipt.id,
        { actorId: session.actorId, reason: t('sco.cancel.reason') },
        {
          headers: createIdempotencyHeaders(createIdempotencyKey('sco-terminal', 'cancel-receipt')),
        },
      );
      setReceipt(null);
      setBarcode(terminalConfig.demoBarcode);
      setStage((prev) => reduceScoStage(prev, { type: 'cancel' }));
    } catch (error) {
      console.error('[sco-terminal] cancel receipt failed', error);
      setErrorMessage(t('sco.errors.cancelFailed'));
    } finally {
      setIsBusy(false);
      setCancelDialogOpen(false);
    }
  }

  function reviewReceipt(): void {
    setStage((prev) => reduceScoStage(prev, { type: 'reviewReceipt' }));
  }

  function resumeScanning(): void {
    setStage((prev) => reduceScoStage(prev, { type: 'resumeScanning' }));
  }

  if (prepStatus === 'error') {
    return <OutOfServiceScreen onRetry={retryPreparation} />;
  }
  if (prepStatus === 'preparing') {
    return <BootScreen messageKey="sco.boot.preparing" />;
  }

  return (
    <main
      className="sco-terminal-shell"
      data-layout={terminalConfig.layoutProfile}
      data-stage={stage}
    >
      <header className="sco-header">
        <div>
          <p className="eyebrow">{t('sco.eyebrow')}</p>
          <h1>{t('sco.appTitle')}</h1>
        </div>
        <div className="sco-header-actions">
          <LanguageSelect />
        </div>
      </header>

      <section
        aria-atomic="true"
        aria-live={errorMessage ? 'assertive' : 'polite'}
        className="sco-status-line"
        role={errorMessage ? 'alert' : 'status'}
      >
        {errorMessage ? (
          <>
            <strong>{errorMessage}</strong>
            <span className="muted">{t('sco.errors.callStaffHint')}</span>
          </>
        ) : (
          <span className="muted">
            {t('sco.footer.store', { storeId: terminalConfig.storeId })} ·{' '}
            {t('sco.footer.terminal', { terminalId: terminalConfig.terminalId })} ·{' '}
            {lastHeartbeatAt ? t('sco.footer.online') : t('sco.footer.offline')}
          </span>
        )}
      </section>

      {stage === 'idle' ? (
        <Card className="sco-idle-card">
          <CardHeading title={t('sco.idle.title')} subtitle={t('sco.idle.subtitle')} />
          <ol className="sco-guidance-list">
            <li>{t('sco.idle.guidanceScan')}</li>
            <li>{t('sco.idle.guidanceCheck')}</li>
            <li>{t('sco.idle.guidancePay')}</li>
          </ol>
          <Badge variant="success">{t('sco.idle.availableBadge')}</Badge>
          <Button type="button" size="lg" disabled={isBusy} onClick={() => void startCheckout()}>
            {t('sco.idle.startAction')}
          </Button>
        </Card>
      ) : (
        <div className="sco-active-grid">
          <Card className="sco-scan-card">
            {stage === 'scanning' ? (
              <>
                <CardHeading
                  title={t('sco.scanning.title')}
                  subtitle={t('sco.scanning.subtitle')}
                />
                <div className="sco-scan-row">
                  <Field>
                    <Label htmlFor="sco-barcode-input">{t('sco.scanning.barcodeLabel')}</Label>
                    <Input
                      id="sco-barcode-input"
                      value={barcode}
                      disabled={isBusy || !canEditReceipt}
                      onChange={(event) => setBarcode(event.target.value)}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter') {
                          void scanBarcode();
                        }
                      }}
                    />
                  </Field>
                  <Button
                    type="button"
                    disabled={isBusy || !canEditReceipt || !barcode.trim()}
                    onClick={() => void scanBarcode()}
                  >
                    {t('sco.scanning.scanAction')}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    disabled={isBusy || !canEditReceipt}
                    onClick={scanDemoBarcode}
                  >
                    {t('sco.scanning.demoBarcodeAction')}
                  </Button>
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  disabled={isBusy || lineCount === 0}
                  onClick={reviewReceipt}
                >
                  {t('sco.scanning.reviewAction')}
                </Button>
              </>
            ) : (
              <>
                <CardHeading
                  title={t('sco.receipt.title')}
                  subtitle={t('sco.receipt.reviewSubtitle')}
                />
                <Button
                  type="button"
                  variant="secondary"
                  disabled={isBusy}
                  onClick={resumeScanning}
                >
                  {t('sco.receipt.backToScanning')}
                </Button>
              </>
            )}

            <div className="sco-stage-indicator" role="list" aria-label={t('sco.stage.label')}>
              {STAGE_INDICATOR_ORDER.map((indicatorStage) => (
                <Badge
                  key={indicatorStage}
                  role="listitem"
                  variant={
                    indicatorStage === stage
                      ? 'accent'
                      : isStageComingSoon(indicatorStage)
                        ? 'outline'
                        : 'default'
                  }
                  aria-current={indicatorStage === stage ? 'step' : undefined}
                >
                  {t(`sco.stage.${indicatorStage}`)}
                  {isStageComingSoon(indicatorStage) ? ` (${t('sco.stage.comingSoon')})` : ''}
                </Badge>
              ))}
            </div>

            <div className="sco-receipt-table-wrap">
              <table className="sco-receipt-table">
                <thead>
                  <tr>
                    <th>{t('sco.receipt.product')}</th>
                    <th>{t('sco.receipt.quantity')}</th>
                    <th>{t('sco.receipt.price')}</th>
                    <th>{t('sco.receipt.total')}</th>
                  </tr>
                </thead>
                <tbody>
                  {!activeReceipt || activeReceipt.lines.length === 0 ? (
                    <tr>
                      <td colSpan={4}>{t('sco.receipt.empty')}</td>
                    </tr>
                  ) : (
                    activeReceipt.lines.map((line) => (
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
            </div>
          </Card>

          <aside className="sco-totals-rail">
            <Card className="sco-total-card sco-total-card--primary">
              <span>{t('sco.receipt.grandTotal')}</span>
              <strong>{formatAmount(totalMinor)}</strong>
            </Card>
            <Card className="sco-total-card">
              <span>{t('sco.receipt.itemCount', { count: lineCount })}</span>
            </Card>
            <Button
              type="button"
              variant="secondary"
              disabled={isBusy || !activeReceipt}
              onClick={() => setCancelDialogOpen(true)}
            >
              {t('sco.cancel.action')}
            </Button>
          </aside>
        </div>
      )}

      <DetailDialog
        open={cancelDialogOpen}
        onOpenChange={setCancelDialogOpen}
        title={t('sco.cancel.confirmTitle')}
        cancelLabel={t('sco.cancel.keepShopping')}
        footer={
          <Button type="button" disabled={isBusy} onClick={() => void confirmCancel()}>
            {t('sco.cancel.confirmAction')}
          </Button>
        }
      >
        <p>{t('sco.cancel.confirmBody')}</p>
      </DetailDialog>
    </main>
  );
}

function AppShell() {
  const { state, retry } = useAuth();
  if (state.status === 'error') {
    return <OutOfServiceScreen onRetry={retry} />;
  }
  if (state.status === 'booting') {
    return <BootScreen messageKey="sco.boot.starting" />;
  }
  return <Terminal session={state.session} />;
}

export function Root() {
  return (
    <ThemeProvider defaultTheme={{ surface: 'sco', colorMode: 'light', accentPreset: 'sco' }}>
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            <AppShell />
          </AuthProvider>
        </QueryClientProvider>
      </I18nextProvider>
    </ThemeProvider>
  );
}
