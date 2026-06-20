import {
  Button,
  LayoutGrid,
  ThemePreview,
  type AccentPreset,
  type LayoutGridSpec,
  type Surface,
} from '@mercadia/ui';
import { useTranslation } from 'react-i18next';

type LayoutTemplatePreviewProps = {
  kind: string;
  accentPreset?: string;
  accentColor?: string;
  resolvedAccentColor?: string;
  grid: LayoutGridSpec;
};

export function LayoutTemplatePreview({
  kind,
  accentPreset,
  accentColor,
  resolvedAccentColor,
  grid,
}: LayoutTemplatePreviewProps) {
  const { t } = useTranslation();
  const surface: Surface = kind === 'sco' ? 'sco' : 'terminal';

  return (
    <ThemePreview
      className="panel theme-preview-panel"
      theme={{
        surface,
        colorMode: 'light',
        accentPreset: (accentPreset as AccentPreset | undefined) || undefined,
        accent: accentColor || resolvedAccentColor || undefined,
      }}
    >
      <div className="panel-heading">
        <div>
          <h3>{t('layoutTemplates.previewTitle')}</h3>
          <p className="muted">
            {t('layoutTemplates.previewAccent')}:{' '}
            {resolvedAccentColor ?? accentColor ?? accentPreset}
          </p>
        </div>
        <Button type="button">{t('layoutTemplates.previewAction')}</Button>
      </div>
      <LayoutGrid grid={grid} />
    </ThemePreview>
  );
}
