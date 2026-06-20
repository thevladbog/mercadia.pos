import { Button, Tabs, TabsList, TabsTrigger } from '@mercadia/ui';
import type { LayoutGridCategorySpec, LayoutGridSpec, LayoutGridTileSpec } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

type LayoutGridEditorProps = {
  grid: LayoutGridSpec;
  catalogReady?: boolean;
  knownProductIds?: ReadonlySet<string>;
  onChange: (grid: LayoutGridSpec) => void;
};

const MIN_DIMENSION = 1;
const MAX_DIMENSION = 12;
const ALL_CATEGORIES = '__all__';

function clampDimension(value: number): number {
  return Math.min(MAX_DIMENSION, Math.max(MIN_DIMENSION, value));
}

function updateTile(
  tiles: LayoutGridTileSpec[],
  index: number,
  patch: Partial<LayoutGridTileSpec>,
) {
  return tiles.map((tile, tileIndex) => (tileIndex === index ? { ...tile, ...patch } : tile));
}

function updateCategory(
  categories: LayoutGridCategorySpec[],
  index: number,
  patch: Partial<LayoutGridCategorySpec>,
) {
  return categories.map((category, categoryIndex) =>
    categoryIndex === index ? { ...category, ...patch } : category,
  );
}

function createCategoryId(): string {
  return `cat-${Date.now().toString(36)}`;
}

export function LayoutGridEditor({
  grid,
  catalogReady,
  knownProductIds,
  onChange,
}: LayoutGridEditorProps) {
  const { t } = useTranslation();
  const [activeCategoryId, setActiveCategoryId] = useState(ALL_CATEGORIES);
  const categories = grid.categories ?? [];

  const visibleTileEntries = useMemo(() => {
    return grid.tiles
      .map((tile, index) => ({ tile, index }))
      .filter(
        ({ tile }) => activeCategoryId === ALL_CATEGORIES || tile.categoryId === activeCategoryId,
      );
  }, [activeCategoryId, grid.tiles]);

  function handleRowsChange(rawValue: string) {
    const parsed = Number.parseInt(rawValue, 10);
    if (Number.isNaN(parsed)) {
      return;
    }
    onChange({ ...grid, rows: clampDimension(parsed) });
  }

  function handleColsChange(rawValue: string) {
    const parsed = Number.parseInt(rawValue, 10);
    if (Number.isNaN(parsed)) {
      return;
    }
    onChange({ ...grid, cols: clampDimension(parsed) });
  }

  function handleAddCategory() {
    const category = { id: createCategoryId(), label: t('layoutTemplates.gridEditor.newCategory') };
    onChange({ ...grid, categories: [...categories, category] });
    setActiveCategoryId(category.id);
  }

  function handleRemoveCategory(categoryId: string) {
    onChange({
      ...grid,
      categories: categories.filter((category) => category.id !== categoryId),
      tiles: grid.tiles.map((tile) =>
        tile.categoryId === categoryId ? { ...tile, categoryId: undefined } : tile,
      ),
    });
    if (activeCategoryId === categoryId) {
      setActiveCategoryId(ALL_CATEGORIES);
    }
  }

  function handleAddTile() {
    onChange({
      ...grid,
      tiles: [
        ...grid.tiles,
        {
          label: '',
          color: '#FF6600',
          ...(activeCategoryId !== ALL_CATEGORIES ? { categoryId: activeCategoryId } : {}),
        },
      ],
    });
  }

  function handleRemoveTile(index: number) {
    onChange({
      ...grid,
      tiles: grid.tiles.filter((_, tileIndex) => tileIndex !== index),
    });
  }

  function isUnknownProduct(productId: string | undefined): boolean {
    if (!productId || !knownProductIds || !catalogReady) {
      return false;
    }
    return !knownProductIds.has(productId);
  }

  return (
    <div className="stack layout-grid-editor">
      <div className="layout-grid-editor-dimensions">
        <label className="field">
          <span>{t('layoutTemplates.gridEditor.rows')}</span>
          <input
            min={MIN_DIMENSION}
            max={MAX_DIMENSION}
            required
            type="number"
            value={grid.rows}
            onChange={(event) => handleRowsChange(event.target.value)}
          />
        </label>
        <label className="field">
          <span>{t('layoutTemplates.gridEditor.cols')}</span>
          <input
            min={MIN_DIMENSION}
            max={MAX_DIMENSION}
            required
            type="number"
            value={grid.cols}
            onChange={(event) => handleColsChange(event.target.value)}
          />
        </label>
      </div>

      <div className="stack">
        <div className="layout-grid-editor-tiles-header">
          <h3>{t('layoutTemplates.gridEditor.categories')}</h3>
          <Button type="button" variant="secondary" onClick={handleAddCategory}>
            {t('layoutTemplates.gridEditor.addCategory')}
          </Button>
        </div>
        {categories.length === 0 ? (
          <p className="muted">{t('layoutTemplates.gridEditor.noCategories')}</p>
        ) : (
          categories.map((category, index) => (
            <div key={category.id} className="layout-grid-editor-tile panel">
              <div className="layout-grid-editor-tile-header">
                <strong>
                  {t('layoutTemplates.gridEditor.categoryNumber', { number: index + 1 })}
                </strong>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => handleRemoveCategory(category.id)}
                >
                  {t('layoutTemplates.gridEditor.removeCategory')}
                </Button>
              </div>
              <label className="field">
                <span>{t('layoutTemplates.gridEditor.categoryLabel')}</span>
                <input
                  value={category.label}
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      categories: updateCategory(categories, index, { label: event.target.value }),
                    })
                  }
                />
              </label>
            </div>
          ))
        )}
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

      <div className="stack">
        <div className="layout-grid-editor-tiles-header">
          <h3>{t('layoutTemplates.gridEditor.tiles')}</h3>
          <Button type="button" variant="secondary" onClick={handleAddTile}>
            {t('layoutTemplates.gridEditor.addTile')}
          </Button>
        </div>

        {visibleTileEntries.length === 0 ? (
          <p className="muted">{t('layoutTemplates.gridEditor.noTiles')}</p>
        ) : (
          visibleTileEntries.map(({ tile, index }) => (
            <div key={index} className="layout-grid-editor-tile panel">
              <div className="layout-grid-editor-tile-header">
                <strong>{t('layoutTemplates.gridEditor.tileNumber', { number: index + 1 })}</strong>
                <Button type="button" variant="ghost" onClick={() => handleRemoveTile(index)}>
                  {t('layoutTemplates.gridEditor.removeTile')}
                </Button>
              </div>
              {categories.length > 0 ? (
                <label className="field">
                  <span>{t('layoutTemplates.gridEditor.category')}</span>
                  <select
                    value={tile.categoryId ?? ''}
                    onChange={(event) =>
                      onChange({
                        ...grid,
                        tiles: updateTile(grid.tiles, index, {
                          categoryId: event.target.value || undefined,
                        }),
                      })
                    }
                  >
                    <option value="">{t('layoutTemplates.gridEditor.uncategorized')}</option>
                    {categories.map((category) => (
                      <option key={category.id} value={category.id}>
                        {category.label}
                      </option>
                    ))}
                  </select>
                </label>
              ) : null}
              <label className="field">
                <span>{t('layoutTemplates.gridEditor.label')}</span>
                <input
                  value={tile.label}
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, { label: event.target.value }),
                    })
                  }
                />
              </label>
              <label className="field">
                <span>{t('layoutTemplates.gridEditor.color')}</span>
                <input
                  placeholder="#FF6600"
                  value={tile.color ?? ''}
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        color: event.target.value || undefined,
                      }),
                    })
                  }
                />
              </label>
              <label className="field">
                <span>{t('layoutTemplates.gridEditor.iconUrl')}</span>
                <input
                  placeholder={t('layoutTemplates.gridEditor.iconUrlPlaceholder')}
                  value={tile.iconUrl ?? ''}
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        iconUrl: event.target.value || undefined,
                      }),
                    })
                  }
                />
              </label>
              <label className="field">
                <span>{t('layoutTemplates.gridEditor.productId')}</span>
                <input
                  value={tile.productId ?? ''}
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        productId: event.target.value || undefined,
                      }),
                    })
                  }
                />
                {isUnknownProduct(tile.productId) ? (
                  <span className="error">
                    {t('layoutTemplates.unknownProduct', { productId: tile.productId })}
                  </span>
                ) : null}
              </label>
              <label className="field layout-grid-editor-checkbox">
                <input
                  checked={tile.empty ?? false}
                  type="checkbox"
                  onChange={(event) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        empty: event.target.checked || undefined,
                      }),
                    })
                  }
                />
                <span>{t('layoutTemplates.gridEditor.empty')}</span>
              </label>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
