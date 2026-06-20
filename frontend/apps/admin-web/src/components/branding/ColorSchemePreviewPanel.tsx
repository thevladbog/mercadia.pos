import {
  Button,
  LayoutGrid,
  Stepper,
  ThemePreview,
  type AccentPreset,
  type LayoutGridSpec,
  type Surface,
} from '@mercadia/ui';
import { useTranslation } from 'react-i18next';

type ColorSchemePreviewPanelProps = {
  accentPreset?: string;
  accentColor?: string;
  backgroundColor?: string;
  surface?: Surface;
};

const demoGrid: LayoutGridSpec = {
  rows: 2,
  cols: 2,
  tiles: [{ label: 'Coffee' }, { label: 'Tea' }, { label: 'Water' }, { empty: true, label: '' }],
};

export function ColorSchemePreviewPanel({
  accentPreset,
  accentColor,
  backgroundColor,
  surface = 'terminal',
}: ColorSchemePreviewPanelProps) {
  const { t } = useTranslation();

  return (
    <ThemePreview
      className="panel theme-preview-panel"
      theme={{
        surface,
        colorMode: 'light',
        accentPreset: (accentPreset as AccentPreset | undefined) || 'neutral',
        accent: accentColor || undefined,
      }}
    >
      <h3>{t('colorSchemes.previewTitle')}</h3>
      <p className="muted">{t('colorSchemes.previewSubtitle')}</p>
      <div
        className="theme-preview-canvas"
        style={
          backgroundColor
            ? { background: backgroundColor, padding: '1rem', borderRadius: '0.75rem' }
            : undefined
        }
      >
        <div className="stack">
          <Button type="button">{t('colorSchemes.previewPrimary')}</Button>
          <Button type="button" variant="secondary">
            {t('colorSchemes.previewSecondary')}
          </Button>
          <LayoutGrid grid={demoGrid} />
          <Stepper value={1} onChange={() => undefined} />
        </div>
      </div>
    </ThemePreview>
  );
}
