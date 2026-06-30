import { useListStores } from '@mercadia/api-clients-central';
import {
  clearSessionToken as clearStoreEdgeSessionToken,
  createAuthSession,
  getSessionToken as getStoreEdgeSessionToken,
  setSessionToken as setStoreEdgeSessionToken,
  setStoreAuthSettings,
  useGetStoreAuthSettings,
  type SetStoreAuthSettingsBody,
} from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { TextField } from '@/components/FormControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { createIdempotencyHeaders } from '@/pages/cash-mutation-utils.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';

type SettingsDraft = {
  failedAttemptLimit: string;
  lockoutDurationSeconds: string;
  posAutoLockSeconds: string;
};

const DEFAULT_DRAFT: SettingsDraft = {
  failedAttemptLimit: '5',
  lockoutDurationSeconds: '900',
  posAutoLockSeconds: '300',
};

function parsePositiveInteger(value: string): number | null {
  const trimmed = value.trim();
  if (!/^\d+$/.test(trimmed)) return null;
  const parsed = Number(trimmed);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : null;
}

function draftToBody(draft: SettingsDraft): SetStoreAuthSettingsBody | null {
  const failedAttemptLimit = parsePositiveInteger(draft.failedAttemptLimit);
  const lockoutDurationSeconds = parsePositiveInteger(draft.lockoutDurationSeconds);
  const posAutoLockSeconds = parsePositiveInteger(draft.posAutoLockSeconds);
  if (
    failedAttemptLimit === null ||
    failedAttemptLimit < 1 ||
    failedAttemptLimit > 20 ||
    lockoutDurationSeconds === null ||
    lockoutDurationSeconds < 60 ||
    lockoutDurationSeconds > 86400 ||
    posAutoLockSeconds === null ||
    posAutoLockSeconds < 30 ||
    posAutoLockSeconds > 86400
  ) {
    return null;
  }
  return { failedAttemptLimit, lockoutDurationSeconds, posAutoLockSeconds };
}

function settingsToDraft(settings: {
  failedAttemptLimit: number;
  lockoutDurationSeconds: number;
  posAutoLockSeconds: number;
}): SettingsDraft {
  return {
    failedAttemptLimit: String(settings.failedAttemptLimit),
    lockoutDurationSeconds: String(settings.lockoutDurationSeconds),
    posAutoLockSeconds: String(settings.posAutoLockSeconds),
  };
}

export function StoreSettingsPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [managerActorId, setManagerActorId] = useState('admin-1');
  const [managerPin, setManagerPin] = useState('');
  const [hasManagerSession, setHasManagerSession] = useState(
    () => getStoreEdgeSessionToken() !== null,
  );
  const [draft, setDraft] = useState<SettingsDraft | null>(null);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const settingsQuery = useGetStoreAuthSettings(activeStoreId, {
    query: { enabled: activeStoreId.length > 0 },
  });
  const settings = settingsQuery.data?.status === 200 ? settingsQuery.data.data.settings : null;
  const loadError = settingsQuery.error != null ? getApiErrorMessage(settingsQuery.error) : null;
  const formDraft = draft ?? (settings ? settingsToDraft(settings) : DEFAULT_DRAFT);

  async function loginManager(): Promise<void> {
    setIsSubmitting(true);
    setError('');
    setMessage('');
    try {
      const response = await createAuthSession({
        actorId: managerActorId.trim(),
        pin: managerPin.trim(),
        storeId: activeStoreId,
      });
      if (response.status === 201) {
        setStoreEdgeSessionToken(response.data.session.token);
        setHasManagerSession(true);
        setManagerPin('');
        setMessage(t('settings.managerLoggedIn'));
      }
    } catch (err) {
      setError(getApiErrorMessage(err));
      setHasManagerSession(false);
    } finally {
      setIsSubmitting(false);
    }
  }

  async function saveSettings(): Promise<void> {
    const body = draftToBody(formDraft);
    if (!body) {
      setError(t('settings.invalidNumber'));
      return;
    }

    setIsSubmitting(true);
    setError('');
    setMessage('');
    try {
      const response = await setStoreAuthSettings(activeStoreId, body, {
        headers: createIdempotencyHeaders(),
      });
      if (response.status === 200) {
        await settingsQuery.refetch();
        setDraft(null);
        setMessage(t('settings.saved'));
      }
    } catch (err) {
      setError(getApiErrorMessage(err));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="stack">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('settings.title')}</h2>
            <p className="muted">{t('settings.subtitle')}</p>
          </div>
          <Button
            type="button"
            variant="secondary"
            disabled={settingsQuery.isFetching || activeStoreId.length === 0}
            onClick={() => void settingsQuery.refetch()}
          >
            {settingsQuery.isFetching ? t('common.refreshing') : t('common.refresh')}
          </Button>
        </div>

        <div className="form-grid form-grid--two">
          <StorePicker
            stores={stores}
            value={activeStoreId}
            onChange={(storeId) => {
              setSelectedStoreId(storeId);
              setDraft(null);
              setHasManagerSession(false);
              setManagerPin('');
              clearStoreEdgeSessionToken();
              setMessage('');
              setError('');
            }}
          />
          <TextField
            label={t('settings.managerActorId')}
            value={managerActorId}
            onValueChange={setManagerActorId}
            placeholder={t('settings.managerActorPlaceholder')}
          />
          <TextField
            label={t('settings.managerPin')}
            type="password"
            autoComplete="off"
            value={managerPin}
            onValueChange={setManagerPin}
            placeholder={t('settings.managerPinPlaceholder')}
          />
        </div>

        <div className="header-actions-inline">
          <Button
            type="button"
            disabled={
              isSubmitting ||
              activeStoreId.length === 0 ||
              managerActorId.trim().length === 0 ||
              managerPin.trim().length === 0
            }
            onClick={() => void loginManager()}
          >
            {isSubmitting ? t('common.submitting') : t('settings.managerLogin')}
          </Button>
          {hasManagerSession && <span className="muted">{t('settings.managerSessionReady')}</span>}
        </div>
        <p className="muted">{t('settings.managerLoginHint')}</p>

        {loadError && <p className="error">{loadError}</p>}
        {!activeStoreId && <p className="muted">{t('common.selectStore')}</p>}
        {message && <p className="muted">{message}</p>}
        {error && <p className="error">{error}</p>}
      </div>

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h3>{t('settings.authPolicy')}</h3>
            <p className="muted">{t('settings.authPolicyHint')}</p>
          </div>
        </div>
        <div className="form-grid form-grid--three">
          <TextField
            label={t('settings.failedAttemptLimit')}
            type="number"
            min={1}
            max={20}
            value={formDraft.failedAttemptLimit}
            onValueChange={(value) =>
              setDraft((current) => ({ ...(current ?? formDraft), failedAttemptLimit: value }))
            }
          />
          <TextField
            label={t('settings.lockoutDurationSeconds')}
            type="number"
            min={60}
            max={86400}
            value={formDraft.lockoutDurationSeconds}
            onValueChange={(value) =>
              setDraft((current) => ({ ...(current ?? formDraft), lockoutDurationSeconds: value }))
            }
          />
          <TextField
            label={t('settings.posAutoLockSeconds')}
            type="number"
            min={30}
            max={86400}
            value={formDraft.posAutoLockSeconds}
            onValueChange={(value) =>
              setDraft((current) => ({ ...(current ?? formDraft), posAutoLockSeconds: value }))
            }
          />
        </div>
        <p className="muted">{t('settings.rangeHint')}</p>
        {settings?.updatedById && (
          <p className="muted">
            {t('settings.updatedBy', {
              actorId: settings.updatedById,
              timestamp: settings.updatedAt ?? t('common.emDash'),
            })}
          </p>
        )}
        <div className="header-actions-inline">
          <Button
            type="button"
            disabled={isSubmitting || !hasManagerSession || activeStoreId.length === 0}
            onClick={() => void saveSettings()}
          >
            {isSubmitting ? t('common.submitting') : t('settings.save')}
          </Button>
        </div>
      </div>
    </section>
  );
}
