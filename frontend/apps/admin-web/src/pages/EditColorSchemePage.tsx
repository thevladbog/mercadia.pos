import {
  getGetColorSchemeQueryKey,
  getListColorSchemesQueryKey,
  useGetColorScheme,
  useListStores,
  useUpdateColorScheme,
  type GetColorScheme200Scheme,
  type UpdateColorSchemeBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';

import { ColorSchemePreviewPanel } from '@/components/branding/ColorSchemePreviewPanel.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { CheckboxField, SelectField, TextField } from '@/components/FormControls.js';
import {
  ACCENT_PRESET_OPTIONS,
  accentPresetLabel,
  BrandingBackLink,
} from '@/pages/branding-shared.js';

type EditColorSchemeFormProps = {
  scheme: GetColorScheme200Scheme;
  schemeId: string;
};

function EditColorSchemeForm({ scheme, schemeId }: EditColorSchemeFormProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];

  const [name, setName] = useState(scheme.name);
  const [logoUrl, setLogoUrl] = useState(scheme.logoUrl ?? '');
  const [accentPreset, setAccentPreset] = useState(scheme.accentPreset ?? 'neutral');
  const [accentColor, setAccentColor] = useState(scheme.accentColor ?? '');
  const [backgroundColor, setBackgroundColor] = useState(scheme.backgroundColor ?? '');
  const [status, setStatus] = useState(scheme.status);
  const [storeIds, setStoreIds] = useState<string[]>(scheme.storeIds ?? []);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useUpdateColorScheme({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 200) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListColorSchemesQueryKey() });
        await queryClient.invalidateQueries({ queryKey: getGetColorSchemeQueryKey(schemeId) });
        void navigate('/central/color-schemes');
      },
      onError: (error) => {
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  function toggleStore(id: string, checked: boolean) {
    if (checked) {
      setStoreIds([...new Set([...storeIds, id])]);
      return;
    }
    setStoreIds(storeIds.filter((value) => value !== id));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    const payload: UpdateColorSchemeBody = {
      name,
      logoUrl,
      status,
      storeIds,
      accentPreset,
      accentColor,
      backgroundColor,
    };

    mutation.mutate({ schemeId, data: payload });
  }

  return (
    <div className="stack">
      <form className="stack" onSubmit={handleSubmit}>
        <p className="readonly-field">{scheme.id}</p>
        <TextField required label={t('colorSchemes.name')} value={name} onValueChange={setName} />
        <TextField label={t('colorSchemes.logoUrl')} value={logoUrl} onValueChange={setLogoUrl} />
        <SelectField
          label={t('colorSchemes.accentPreset')}
          value={accentPreset}
          onValueChange={setAccentPreset}
        >
          {ACCENT_PRESET_OPTIONS.map((preset) => (
            <option key={preset} value={preset}>
              {accentPresetLabel(t, preset)}
            </option>
          ))}
        </SelectField>
        <TextField
          label={t('colorSchemes.accentColor')}
          value={accentColor}
          onValueChange={setAccentColor}
        />
        <TextField
          label={t('colorSchemes.backgroundColor')}
          value={backgroundColor}
          onValueChange={setBackgroundColor}
        />
        <SelectField label={t('colorSchemes.status')} value={status} onValueChange={setStatus}>
          <option value="draft">{t('branding.status.draft')}</option>
          <option value="published">{t('branding.status.published')}</option>
        </SelectField>
        {stores.length > 0 ? (
          <fieldset className="role-fieldset">
            <legend>{t('colorSchemes.storeIds')}</legend>
            <div className="role-options">
              {stores.map((store) => (
                <CheckboxField
                  checked={storeIds.includes(store.id)}
                  key={store.id}
                  label={`${store.name} (${store.id})`}
                  onCheckedChange={(checked) => toggleStore(store.id, checked)}
                />
              ))}
            </div>
          </fieldset>
        ) : null}
        {errorMessage ? <p className="error">{errorMessage}</p> : null}
        <div className="form-actions">
          <Button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? t('colorSchemes.saving') : t('common.save')}
          </Button>
        </div>
      </form>
      <ColorSchemePreviewPanel
        accentColor={accentColor}
        accentPreset={accentPreset}
        backgroundColor={backgroundColor}
      />
    </div>
  );
}

export function EditColorSchemePage() {
  const { t } = useTranslation();
  const { schemeId = '' } = useParams();
  const schemeQuery = useGetColorScheme(schemeId, { query: { enabled: schemeId.length > 0 } });
  const scheme = schemeQuery.data?.status === 200 ? schemeQuery.data.data.scheme : null;
  const loadError = schemeQuery.error != null ? getApiErrorMessage(schemeQuery.error) : null;

  return (
    <section className="stack users-page">
      <BrandingBackLink label={t('colorSchemes.backToList')} to="/central/color-schemes" />
      <div className="panel">
        <h2>{t('colorSchemes.editTitle')}</h2>
        {loadError ? (
          <p className="error">{loadError}</p>
        ) : schemeQuery.isLoading && !scheme ? (
          <p className="muted">{t('colorSchemes.loading')}</p>
        ) : scheme ? (
          <EditColorSchemeForm key={scheme.id} scheme={scheme} schemeId={schemeId} />
        ) : null}
      </div>
    </section>
  );
}
