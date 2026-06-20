import { QueryClientProvider } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';
import { useGetLayoutTemplate } from '@mercadia/api-clients-central';
import {
  applyTheme,
  Button,
  LayoutGrid,
  Numpad,
  Tabs,
  TabsList,
  TabsTrigger,
  ThemeProvider,
  type AccentPreset,
  type LayoutGridSpec,
} from '@mercadia/ui';

import { queryClient } from '@/query-client.js';

const ALL_CATEGORIES = '__all__';

function resolveTemplateId(): string {
  const params = new URLSearchParams(window.location.search);
  return params.get('templateId') ?? import.meta.env.VITE_LAYOUT_TEMPLATE_ID ?? '';
}

function filterGridByCategory(grid: LayoutGridSpec, categoryId: string | null): LayoutGridSpec {
  if (categoryId === null) {
    return grid;
  }
  return {
    ...grid,
    tiles: grid.tiles.filter((tile) => tile.categoryId === categoryId),
  };
}

function TerminalShell() {
  const templateId = useMemo(() => resolveTemplateId(), []);
  const templateQuery = useGetLayoutTemplate(templateId, {
    query: { enabled: templateId.length > 0 },
  });
  const template = templateQuery.data?.status === 200 ? templateQuery.data.data.template : null;
  const [numpadValue, setNumpadValue] = useState('');
  const [activeCategoryId, setActiveCategoryId] = useState(ALL_CATEGORIES);

  useEffect(() => {
    if (!template) {
      return;
    }
    applyTheme({
      surface: template.kind === 'sco' ? 'sco' : 'terminal',
      colorMode: 'light',
      accentPreset: template.resolvedAccentPreset as AccentPreset,
      accent: template.resolvedAccentColor,
    });
  }, [template]);

  const grid: LayoutGridSpec = template
    ? {
        rows: template.grid.rows ?? 4,
        cols: template.grid.cols ?? 4,
        categories: (template.grid.categories ?? [])
          .filter((category) => category.id)
          .map((category) => ({
            id: category.id!,
            label: category.label ?? '',
          })),
        tiles: (template.grid.tiles ?? []).map((tile) => ({
          label: tile.label ?? '',
          color: tile.color,
          productId: tile.productId,
          empty: tile.empty,
          categoryId: tile.categoryId,
          iconUrl: tile.iconUrl,
        })),
      }
    : { rows: 4, cols: 4, tiles: [{ label: 'Demo item' }, { label: 'Return item' }] };

  const categories = grid.categories ?? [];
  const resolvedCategoryId = useMemo(() => {
    if (activeCategoryId === ALL_CATEGORIES) {
      return ALL_CATEGORIES;
    }
    return categories.some((category) => category.id === activeCategoryId)
      ? activeCategoryId
      : ALL_CATEGORIES;
  }, [activeCategoryId, categories]);
  const displayGrid = useMemo(
    () =>
      filterGridByCategory(grid, resolvedCategoryId === ALL_CATEGORIES ? null : resolvedCategoryId),
    [grid, resolvedCategoryId],
  );

  return (
    <main className="pos-terminal-shell">
      <header className="pos-terminal-header">
        <h1>{template?.name ?? 'POS Terminal (dev)'}</h1>
        <p className="muted">
          {template
            ? `${template.kind} · ${template.resolvedAccentColor}`
            : 'Add ?templateId=... or VITE_LAYOUT_TEMPLATE_ID'}
        </p>
      </header>
      {!templateId ? (
        <p className="muted">No layout template selected.</p>
      ) : templateQuery.isLoading ? (
        <p className="muted">Loading template…</p>
      ) : templateQuery.isError ? (
        <p className="error">
          Failed to load template. Set VITE_CENTRAL_SESSION_TOKEN for API access.
        </p>
      ) : (
        <div className="pos-terminal-grid">
          <section className="panel">
            <Button type="button">Start sale</Button>
            {categories.length > 0 ? (
              <Tabs value={resolvedCategoryId} onValueChange={setActiveCategoryId}>
                <TabsList aria-label="Categories">
                  <TabsTrigger value={ALL_CATEGORIES}>All</TabsTrigger>
                  {categories.map((category) => (
                    <TabsTrigger key={category.id} value={category.id}>
                      {category.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            ) : null}
            <LayoutGrid grid={displayGrid} onTileClick={() => undefined} />
          </section>
          <section className="panel">
            <Numpad enterLabel="Enter" value={numpadValue} onChange={setNumpadValue} />
          </section>
        </div>
      )}
    </main>
  );
}

export function Root() {
  return (
    <ThemeProvider defaultTheme={{ surface: 'terminal', colorMode: 'light', accentPreset: 'sale' }}>
      <QueryClientProvider client={queryClient}>
        <TerminalShell />
      </QueryClientProvider>
    </ThemeProvider>
  );
}
