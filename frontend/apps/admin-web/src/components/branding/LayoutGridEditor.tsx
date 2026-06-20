import { Button, Tabs, TabsList, TabsTrigger } from '@mercadia/ui';
import type { LayoutGridCategorySpec, LayoutGridSpec, LayoutGridTileSpec } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { CheckboxField, SelectField, TextField } from '@/components/FormControls.js';

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

  function isUnknownProduct(productId: string | undefined, empty?: boolean): boolean {
    if (!productId || empty || !knownProductIds || !catalogReady) {
      return false;
    }
    return !knownProductIds.has(productId);
  }

  return (
    <div className="stack layout-grid-editor">
      <div className="layout-grid-editor-dimensions">
        <TextField
          required
          label={t('layoutTemplates.gridEditor.rows')}
          max={MAX_DIMENSION}
          min={MIN_DIMENSION}
          type="number"
          value={grid.rows}
          onValueChange={handleRowsChange}
        />
        <TextField
          required
          label={t('layoutTemplates.gridEditor.cols')}
          max={MAX_DIMENSION}
          min={MIN_DIMENSION}
          type="number"
          value={grid.cols}
          onValueChange={handleColsChange}
        />
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
              <TextField
                label={t('layoutTemplates.gridEditor.categoryLabel')}
                value={category.label}
                onValueChange={(value) =>
                  onChange({
                    ...grid,
                    categories: updateCategory(categories, index, { label: value }),
                  })
                }
              />
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
                <SelectField
                  label={t('layoutTemplates.gridEditor.category')}
                  value={tile.categoryId ?? ''}
                  onValueChange={(value) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        categoryId: value || undefined,
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
                </SelectField>
              ) : null}
              <TextField
                label={t('layoutTemplates.gridEditor.label')}
                value={tile.label}
                onValueChange={(value) =>
                  onChange({
                    ...grid,
                    tiles: updateTile(grid.tiles, index, { label: value }),
                  })
                }
              />
              <TextField
                label={t('layoutTemplates.gridEditor.color')}
                placeholder="#FF6600"
                value={tile.color ?? ''}
                onValueChange={(value) =>
                  onChange({
                    ...grid,
                    tiles: updateTile(grid.tiles, index, {
                      color: value || undefined,
                    }),
                  })
                }
              />
              <TextField
                label={t('layoutTemplates.gridEditor.iconUrl')}
                placeholder={t('layoutTemplates.gridEditor.iconUrlPlaceholder')}
                value={tile.iconUrl ?? ''}
                onValueChange={(value) =>
                  onChange({
                    ...grid,
                    tiles: updateTile(grid.tiles, index, {
                      iconUrl: value || undefined,
                    }),
                  })
                }
              />
              <div className="field">
                <TextField
                  label={t('layoutTemplates.gridEditor.productId')}
                  value={tile.productId ?? ''}
                  onValueChange={(value) =>
                    onChange({
                      ...grid,
                      tiles: updateTile(grid.tiles, index, {
                        productId: value || undefined,
                      }),
                    })
                  }
                />
                {isUnknownProduct(tile.productId, tile.empty) ? (
                  <span className="error">
                    {t('layoutTemplates.unknownProduct', { productId: tile.productId })}
                  </span>
                ) : null}
              </div>
              <CheckboxField
                checked={tile.empty ?? false}
                className="layout-grid-editor-checkbox"
                label={t('layoutTemplates.gridEditor.empty')}
                onCheckedChange={(checked) =>
                  onChange({
                    ...grid,
                    tiles: updateTile(grid.tiles, index, {
                      empty: checked || undefined,
                    }),
                  })
                }
              />
            </div>
          ))
        )}
      </div>
    </div>
  );
}
