import type { Meta, StoryObj } from '@storybook/react-vite';
import type { CSSProperties, ReactNode } from 'react';
import { useEffect, useState } from 'react';

import { Badge, Button, Card, CardHeading, ThemePreview } from '../index.js';
import type { AccentPreset, ColorMode, Surface } from '../index.js';
import { getStorybookLocale, type StorybookLocale } from './locale.js';

const foundationCopy = {
  en: {
    toolbarHint: 'Use the Storybook toolbar to switch surface, mode, accent, and language.',
    primitiveColors: 'Primitive colors',
    semanticColors: 'Semantic colors',
    spacingAndLayout: 'Spacing and layout',
    radiiAndShadows: 'Radii and shadows',
    surfaceTokens: 'Surface tokens',
    themeMatrixSubtitle: 'Surface preset across modes and accents',
    operationalSample: 'Operational UI sample',
    primaryAction: 'Primary action',
  },
  ru: {
    toolbarHint: 'Используйте toolbar Storybook для переключения surface, mode, accent и языка.',
    primitiveColors: 'Базовые цвета',
    semanticColors: 'Семантические цвета',
    spacingAndLayout: 'Отступы и размеры',
    radiiAndShadows: 'Радиусы и тени',
    surfaceTokens: 'Переменные surface',
    themeMatrixSubtitle: 'Surface preset в разных режимах и акцентах',
    operationalSample: 'Пример операционного интерфейса',
    primaryAction: 'Основное действие',
  },
};

type Token = {
  name: string;
  description?: string;
  preview: 'color' | 'size' | 'radius' | 'shadow' | 'text';
};

const primitiveColorTokens: Token[] = [
  { name: '--primitive-brand-orange', preview: 'color' },
  { name: '--primitive-blue-600', preview: 'color' },
  { name: '--primitive-sco-orange', preview: 'color' },
  { name: '--primitive-gray-50', preview: 'color' },
  { name: '--primitive-gray-100', preview: 'color' },
  { name: '--primitive-gray-200', preview: 'color' },
  { name: '--primitive-gray-400', preview: 'color' },
  { name: '--primitive-gray-500', preview: 'color' },
  { name: '--primitive-gray-700', preview: 'color' },
  { name: '--primitive-gray-900', preview: 'color' },
  { name: '--primitive-charcoal', preview: 'color' },
  { name: '--primitive-white', preview: 'color' },
  { name: '--primitive-success', preview: 'color' },
  { name: '--primitive-warning', preview: 'color' },
  { name: '--primitive-danger', preview: 'color' },
  { name: '--primitive-info', preview: 'color' },
];

const semanticColorTokens: Token[] = [
  { name: '--ui-bg', preview: 'color' },
  { name: '--ui-surface', preview: 'color' },
  { name: '--ui-surface-elevated', preview: 'color' },
  { name: '--ui-surface-dark', preview: 'color' },
  { name: '--ui-text', preview: 'color' },
  { name: '--ui-text-muted', preview: 'color' },
  { name: '--ui-border', preview: 'color' },
  { name: '--ui-overlay', preview: 'color' },
  { name: '--ui-accent', preview: 'color' },
  { name: '--ui-accent-hover', preview: 'color' },
  { name: '--ui-accent-muted', preview: 'color' },
  { name: '--ui-accent-foreground', preview: 'color' },
  { name: '--ui-success', preview: 'color' },
  { name: '--ui-success-muted', preview: 'color' },
  { name: '--ui-warning', preview: 'color' },
  { name: '--ui-warning-muted', preview: 'color' },
  { name: '--ui-danger', preview: 'color' },
  { name: '--ui-danger-muted', preview: 'color' },
  { name: '--ui-info', preview: 'color' },
  { name: '--ui-info-muted', preview: 'color' },
];

const spacingTokens: Token[] = [
  { name: '--ui-space-xs', preview: 'size' },
  { name: '--ui-space-sm', preview: 'size' },
  { name: '--ui-space-md', preview: 'size' },
  { name: '--ui-space-lg', preview: 'size' },
  { name: '--ui-space-xl', preview: 'size' },
  { name: '--ui-touch-min', preview: 'size' },
  { name: '--ui-sidebar-width', preview: 'size' },
];

const radiusTokens: Token[] = [
  { name: '--ui-radius-sm', preview: 'radius' },
  { name: '--ui-radius-md', preview: 'radius' },
  { name: '--ui-radius-lg', preview: 'radius' },
  { name: '--ui-radius-pill', preview: 'radius' },
];

const shadowTokens: Token[] = [
  { name: '--ui-shadow-card', preview: 'shadow' },
  { name: '--ui-shadow-dropdown', preview: 'shadow' },
];

const surfaceTokens: Token[] = [
  { name: '--ui-density', preview: 'text' },
  { name: '--ui-sidebar-width', preview: 'size' },
  { name: '--ui-touch-min', preview: 'size' },
];

const surfaces: Surface[] = ['admin', 'terminal', 'sco', 'senior-cashier'];
const colorModes: ColorMode[] = ['light', 'dark'];
const accentPresets: AccentPreset[] = ['neutral', 'sale', 'return', 'sco'];

function useCssVariable(name: string) {
  const [value, setValue] = useState('');

  useEffect(() => {
    function updateValue() {
      setValue(getComputedStyle(document.documentElement).getPropertyValue(name).trim());
    }

    updateValue();
    const frame = window.requestAnimationFrame(updateValue);
    return () => window.cancelAnimationFrame(frame);
  }, [name]);

  return value;
}

function TokenPreview({ token, value }: { token: Token; value: string }) {
  const previewValue = value || 'transparent';
  const style = { '--token-preview': previewValue } as CSSProperties;

  if (token.preview === 'color') {
    return <div className="mercadia-token-swatch" style={style} />;
  }
  if (token.preview === 'size') {
    return <div className="mercadia-token-size-preview" style={style} />;
  }
  if (token.preview === 'radius') {
    return <div className="mercadia-token-radius-preview" style={style} />;
  }
  if (token.preview === 'shadow') {
    return <div className="mercadia-token-shadow-preview" style={style} />;
  }
  return <span className="mercadia-token-value">{value || '(not set)'}</span>;
}

function TokenRow({ token }: { token: Token }) {
  const value = useCssVariable(token.name);

  return (
    <div className="mercadia-token-row">
      <div>
        <div className="mercadia-token-name">{token.name}</div>
        {token.description ? <div className="mercadia-token-value">{token.description}</div> : null}
      </div>
      <div className="mercadia-token-value">{value || '(not set)'}</div>
      <TokenPreview token={token} value={value} />
    </div>
  );
}

function TokenTable({ tokens }: { tokens: Token[] }) {
  return (
    <div className="mercadia-token-table">
      {tokens.map((token) => (
        <TokenRow key={token.name} token={token} />
      ))}
    </div>
  );
}

function Section({
  title,
  locale,
  children,
}: {
  title: string;
  locale: StorybookLocale;
  children: ReactNode;
}) {
  return (
    <div className="mercadia-story-section">
      <Card>
        <CardHeading title={title} subtitle={foundationCopy[locale].toolbarHint} />
      </Card>
      {children}
    </div>
  );
}

const meta = {
  title: 'Foundations/Tokens',
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const PrimitiveColors: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    return (
      <Section title={foundationCopy[locale].primitiveColors} locale={locale}>
        <TokenTable tokens={primitiveColorTokens} />
      </Section>
    );
  },
};

export const SemanticColors: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    return (
      <Section title={foundationCopy[locale].semanticColors} locale={locale}>
        <TokenTable tokens={semanticColorTokens} />
      </Section>
    );
  },
};

export const SpacingAndLayout: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    return (
      <Section title={foundationCopy[locale].spacingAndLayout} locale={locale}>
        <TokenTable tokens={spacingTokens} />
      </Section>
    );
  },
};

export const RadiiAndShadows: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    return (
      <Section title={foundationCopy[locale].radiiAndShadows} locale={locale}>
        <TokenTable tokens={[...radiusTokens, ...shadowTokens]} />
      </Section>
    );
  },
};

export const SurfaceTokens: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    return (
      <Section title={foundationCopy[locale].surfaceTokens} locale={locale}>
        <TokenTable tokens={surfaceTokens} />
      </Section>
    );
  },
};

export const ThemeMatrix: Story = {
  render: (_args, context) => {
    const locale = getStorybookLocale(context.globals.locale);
    const copy = foundationCopy[locale];
    return (
      <div className="mercadia-story-section">
        {surfaces.map((surface) => (
          <Card key={surface}>
            <CardHeading title={surface} subtitle={copy.themeMatrixSubtitle} />
            <div className="mercadia-story-grid">
              {colorModes.map((colorMode) =>
                accentPresets.map((accentPreset) => (
                  <ThemePreview
                    key={`${surface}-${colorMode}-${accentPreset}`}
                    className="mercadia-theme-sample"
                    theme={{ surface, colorMode, accentPreset }}
                  >
                    <div className="mercadia-story-row">
                      <Badge variant="accent">{accentPreset}</Badge>
                      <Badge variant="outline">{colorMode}</Badge>
                    </div>
                    <strong>{surface}</strong>
                    <span style={{ color: 'var(--ui-text-muted)' }}>{copy.operationalSample}</span>
                    <Button size="sm">{copy.primaryAction}</Button>
                  </ThemePreview>
                )),
              )}
            </div>
          </Card>
        ))}
      </div>
    );
  },
};
