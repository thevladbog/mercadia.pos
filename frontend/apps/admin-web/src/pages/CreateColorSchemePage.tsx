import {
  getListColorSchemesQueryKey,
  useCreateColorScheme,
  useListStores,
  type CreateColorSchemeBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { ColorSchemePreviewPanel } from '@/components/branding/ColorSchemePreviewPanel.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  ACCENT_PRESET_OPTIONS,
  accentPresetLabel,
  BrandingBackLink,
} from '@/pages/branding-shared.js';

export function CreateColorSchemePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];

  const [schemeId, setSchemeId] = useState('');
  const [name, setName] = useState('');
  const [logoUrl, setLogoUrl] = useState('');
  const [accentPreset, setAccentPreset] = useState('neutral');
  const [accentColor, setAccentColor] = useState('');
  const [backgroundColor, setBackgroundColor] = useState('');
  const [status, setStatus] = useState('draft');
  const [storeIds, setStoreIds] = useState<string[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useCreateColorScheme({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 201) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListColorSchemesQueryKey() });
        void navigate('/central/color-schemes');
      },
      onError: (error) => {
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  function toggleStore(storeId: string, checked: boolean) {
    if (checked) {
      setStoreIds([...new Set([...storeIds, storeId])]);
      return;
    }
    setStoreIds(storeIds.filter((id) => id !== storeId));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    const payload: CreateColorSchemeBody = {
      schemeId: schemeId.trim(),
      name: name.trim(),
      status,
      storeIds,
      ...(logoUrl.trim() ? { logoUrl: logoUrl.trim() } : {}),
      ...(accentColor.trim() ? { accentColor: accentColor.trim() } : { accentPreset }),
      ...(backgroundColor.trim() ? { backgroundColor: backgroundColor.trim() } : {}),
    };

    mutation.mutate({ data: payload });
  }

  return (
    <section className="stack users-page">
      <BrandingBackLink label={t('colorSchemes.backToList')} to="/central/color-schemes" />
      <div className="panel">
        <h2>{t('colorSchemes.createTitle')}</h2>
        <div className="stack">
          <form className="stack" onSubmit={handleSubmit}>
            <label className="field">
              <span>{t('colorSchemes.schemeId')}</span>
              <input required value={schemeId} onChange={(e) => setSchemeId(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('colorSchemes.name')}</span>
              <input required value={name} onChange={(e) => setName(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('colorSchemes.logoUrl')}</span>
              <input value={logoUrl} onChange={(e) => setLogoUrl(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('colorSchemes.accentPreset')}</span>
              <select value={accentPreset} onChange={(e) => setAccentPreset(e.target.value)}>
                {ACCENT_PRESET_OPTIONS.map((preset) => (
                  <option key={preset} value={preset}>
                    {accentPresetLabel(t, preset)}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t('colorSchemes.accentColor')}</span>
              <input
                placeholder="#FF6600"
                value={accentColor}
                onChange={(e) => setAccentColor(e.target.value)}
              />
            </label>
            <label className="field">
              <span>{t('colorSchemes.backgroundColor')}</span>
              <input
                placeholder="#F5F5F0"
                value={backgroundColor}
                onChange={(e) => setBackgroundColor(e.target.value)}
              />
            </label>
            <label className="field">
              <span>{t('colorSchemes.status')}</span>
              <select value={status} onChange={(e) => setStatus(e.target.value)}>
                <option value="draft">{t('branding.status.draft')}</option>
                <option value="published">{t('branding.status.published')}</option>
              </select>
            </label>
            {stores.length > 0 ? (
              <fieldset className="role-fieldset">
                <legend>{t('colorSchemes.storeIds')}</legend>
                <div className="role-options">
                  {stores.map((store) => (
                    <label className="checkbox-field" key={store.id}>
                      <input
                        checked={storeIds.includes(store.id)}
                        type="checkbox"
                        onChange={(e) => toggleStore(store.id, e.target.checked)}
                      />
                      <span>
                        {store.name} ({store.id})
                      </span>
                    </label>
                  ))}
                </div>
              </fieldset>
            ) : null}
            {errorMessage ? <p className="error">{errorMessage}</p> : null}
            <div className="form-actions">
              <Button disabled={mutation.isPending} type="submit">
                {mutation.isPending ? t('colorSchemes.creating') : t('colorSchemes.createScheme')}
              </Button>
            </div>
          </form>
          <ColorSchemePreviewPanel
            accentColor={accentColor}
            accentPreset={accentPreset}
            backgroundColor={backgroundColor}
          />
        </div>
      </div>
    </section>
  );
}
