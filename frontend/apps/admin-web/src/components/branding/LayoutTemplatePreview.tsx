import {
  Button,
  LayoutGrid,
  Tabs,
  TabsList,
  TabsTrigger,
  ThemePreview,
  type AccentPreset,
  type LayoutGridSpec,
  type Surface,
} from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { filterGridByCategory } from '@/pages/layout-template-utils.js';

type LayoutTemplatePreviewProps = {
  kind: string;
  accentPreset?: string;
  accentColor?: string;
  resolvedAccentColor?: string;
  grid: LayoutGridSpec;
};

const ALL_CATEGORIES = '__all__';

export function LayoutTemplatePreview({
  kind,
  accentPreset,
  accentColor,
  resolvedAccentColor,
  grid,
}: LayoutTemplatePreviewProps) {
  const { t } = useTranslation();
  const surface: Surface = kind === 'sco' ? 'sco' : 'terminal';
  const categories = grid.categories ?? [];
  const [activeCategoryId, setActiveCategoryId] = useState(ALL_CATEGORIES);
  const previewGrid = useMemo(
    () => filterGridByCategory(grid, activeCategoryId === ALL_CATEGORIES ? null : activeCategoryId),
    [activeCategoryId, grid],
  );

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
      {categories.length > 0 ? (
        <Tabs value={activeCategoryId} onValueChange={setActiveCategoryId}>
          <TabsList aria-label={t('layoutTemplates.gridEditor.categories')}>
            <TabsTrigger value={ALL_CATEGORIES}>
              {t('layoutTemplates.gridEditor.allCategories')}
            </TabsTrigger>
            {categories.map((category) => (
              <TabsTrigger key={category.id} value={category.id}>
                {category.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      ) : null}
      <LayoutGrid grid={previewGrid} />
    </ThemePreview>
  );
}
