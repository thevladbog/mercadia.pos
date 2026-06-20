import {
  getListLayoutTemplatesQueryKey,
  useCreateLayoutTemplate,
  useListColorSchemes,
  useListStores,
  type CreateLayoutTemplateBody,
} from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { LayoutTemplatePreview } from '@/components/branding/LayoutTemplatePreview.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  ACCENT_PRESET_OPTIONS,
  accentPresetLabel,
  BrandingBackLink,
} from '@/pages/branding-shared.js';
import { defaultGrid, gridToApi, parseGridJson } from '@/pages/layout-template-utils.js';

const KIND_OPTIONS = ['sale', 'return', 'sco'] as const;

export function CreateLayoutTemplatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storesQuery = useListStores();
  const schemesQuery = useListColorSchemes();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const schemes = schemesQuery.data?.status === 200 ? schemesQuery.data.data.schemes : [];

  const [templateId, setTemplateId] = useState('');
  const [name, setName] = useState('');
  const [kind, setKind] = useState<string>('sale');
  const [accentPreset, setAccentPreset] = useState('');
  const [accentColor, setAccentColor] = useState('');
  const [colorSchemeId, setColorSchemeId] = useState('');
  const [storeId, setStoreId] = useState('');
  const [terminalType, setTerminalType] = useState('');
  const [status, setStatus] = useState('draft');
  const [gridJson, setGridJson] = useState(JSON.stringify(defaultGrid(), null, 2));
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const grid = useMemo(() => parseGridJson(gridJson) ?? defaultGrid(), [gridJson]);

  const mutation = useCreateLayoutTemplate({
    mutation: {
      onSuccess: async (response) => {
        if (response.status !== 201) {
          setErrorMessage(t('common.unexpectedError'));
          return;
        }
        await queryClient.invalidateQueries({ queryKey: getListLayoutTemplatesQueryKey() });
        void navigate('/central/layout-templates');
      },
      onError: (error) => {
        setErrorMessage(getApiErrorMessage(error));
      },
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    const parsedGrid = parseGridJson(gridJson);
    if (!parsedGrid) {
      setErrorMessage(t('layoutTemplates.invalidGrid'));
      return;
    }

    const payload: CreateLayoutTemplateBody = {
      templateId: templateId.trim(),
      name: name.trim(),
      kind,
      status,
      grid: gridToApi(parsedGrid),
      ...(accentColor.trim()
        ? { accentColor: accentColor.trim() }
        : accentPreset
          ? { accentPreset }
          : {}),
      ...(colorSchemeId ? { colorSchemeId } : {}),
      ...(storeId ? { storeId } : {}),
      ...(terminalType.trim() ? { terminalType: terminalType.trim() } : {}),
    };

    mutation.mutate({ data: payload });
  }

  return (
    <section className="stack users-page">
      <BrandingBackLink label={t('layoutTemplates.backToList')} to="/central/layout-templates" />
      <div className="panel">
        <h2>{t('layoutTemplates.createTitle')}</h2>
        <div className="stack">
          <form className="stack" onSubmit={handleSubmit}>
            <label className="field">
              <span>{t('layoutTemplates.templateId')}</span>
              <input required value={templateId} onChange={(e) => setTemplateId(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('layoutTemplates.name')}</span>
              <input required value={name} onChange={(e) => setName(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('layoutTemplates.kind')}</span>
              <select value={kind} onChange={(e) => setKind(e.target.value)}>
                {KIND_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {t(`layoutTemplates.kind.${option}`)}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t('layoutTemplates.accentPreset')}</span>
              <select value={accentPreset} onChange={(e) => setAccentPreset(e.target.value)}>
                <option value="">{t('layoutTemplates.accentPresetDefault')}</option>
                {ACCENT_PRESET_OPTIONS.map((preset) => (
                  <option key={preset} value={preset}>
                    {accentPresetLabel(t, preset)}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t('layoutTemplates.accentColor')}</span>
              <input value={accentColor} onChange={(e) => setAccentColor(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('layoutTemplates.colorSchemeId')}</span>
              <select value={colorSchemeId} onChange={(e) => setColorSchemeId(e.target.value)}>
                <option value="">{t('layoutTemplates.noColorScheme')}</option>
                {schemes.map((scheme) => (
                  <option key={scheme.id} value={scheme.id}>
                    {scheme.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t('layoutTemplates.storeId')}</span>
              <select value={storeId} onChange={(e) => setStoreId(e.target.value)}>
                <option value="">{t('layoutTemplates.anyStore')}</option>
                {stores.map((store) => (
                  <option key={store.id} value={store.id}>
                    {store.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>{t('layoutTemplates.terminalType')}</span>
              <input value={terminalType} onChange={(e) => setTerminalType(e.target.value)} />
            </label>
            <label className="field">
              <span>{t('layoutTemplates.status')}</span>
              <select value={status} onChange={(e) => setStatus(e.target.value)}>
                <option value="draft">{t('branding.status.draft')}</option>
                <option value="published">{t('branding.status.published')}</option>
              </select>
            </label>
            <label className="field">
              <span>{t('layoutTemplates.gridJson')}</span>
              <textarea rows={8} value={gridJson} onChange={(e) => setGridJson(e.target.value)} />
            </label>
            {errorMessage ? <p className="error">{errorMessage}</p> : null}
            <div className="form-actions">
              <Button disabled={mutation.isPending} type="submit">
                {mutation.isPending
                  ? t('layoutTemplates.creating')
                  : t('layoutTemplates.createTemplate')}
              </Button>
            </div>
          </form>
          <LayoutTemplatePreview
            accentColor={accentColor}
            accentPreset={accentPreset}
            grid={grid}
            kind={kind}
          />
        </div>
      </div>
    </section>
  );
}
