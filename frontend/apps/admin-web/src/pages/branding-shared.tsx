import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

type BrandingBackLinkProps = {
  label?: string;
  to: string;
};

export function BrandingBackLink({ label, to }: BrandingBackLinkProps) {
  const { t } = useTranslation();
  return (
    <p className="page-back">
      <Link to={to}>{label ?? t('common.back')}</Link>
    </p>
  );
}

export const ACCENT_PRESET_OPTIONS = ['sale', 'return', 'sco', 'neutral'] as const;

export function accentPresetLabel(t: (key: string) => string, preset: string): string {
  const key = `branding.accentPreset.${preset}`;
  const translated = t(key);
  return translated === key ? preset : translated;
}

export function statusLabel(t: (key: string) => string, status: string): string {
  const key = `branding.status.${status}`;
  const translated = t(key);
  return translated === key ? status : translated;
}

export function kindLabel(t: (key: string) => string, kind: string): string {
  const key = `layoutTemplates.kind.${kind}`;
  const translated = t(key);
  return translated === key ? kind : translated;
}

export function AccentSwatch({ color }: { color: string }) {
  return (
    <span
      aria-hidden
      style={{
        display: 'inline-block',
        width: '1rem',
        height: '1rem',
        borderRadius: '9999px',
        background: color,
        border: '1px solid var(--ui-border, #e8e8e4)',
        verticalAlign: 'middle',
      }}
    />
  );
}
