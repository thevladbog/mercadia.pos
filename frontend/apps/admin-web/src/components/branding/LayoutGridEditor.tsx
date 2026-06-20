import { Button } from '@mercadia/ui';
import type { LayoutGridSpec, LayoutGridTileSpec } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';

type LayoutGridEditorProps = {
  grid: LayoutGridSpec;
  onChange: (grid: LayoutGridSpec) => void;
};

const MIN_DIMENSION = 1;
const MAX_DIMENSION = 12;

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

export function LayoutGridEditor({ grid, onChange }: LayoutGridEditorProps) {
  const { t } = useTranslation();

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

  function handleAddTile() {
    onChange({
      ...grid,
      tiles: [...grid.tiles, { label: '', color: '#FF6600' }],
    });
  }

  function handleRemoveTile(index: number) {
    onChange({
      ...grid,
      tiles: grid.tiles.filter((_, tileIndex) => tileIndex !== index),
    });
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
          <h3>{t('layoutTemplates.gridEditor.tiles')}</h3>
          <Button type="button" variant="secondary" onClick={handleAddTile}>
            {t('layoutTemplates.gridEditor.addTile')}
          </Button>
        </div>

        {grid.tiles.length === 0 ? (
          <p className="muted">{t('layoutTemplates.gridEditor.noTiles')}</p>
        ) : (
          grid.tiles.map((tile, index) => (
            <div key={index} className="layout-grid-editor-tile panel">
              <div className="layout-grid-editor-tile-header">
                <strong>{t('layoutTemplates.gridEditor.tileNumber', { number: index + 1 })}</strong>
                <Button type="button" variant="ghost" onClick={() => handleRemoveTile(index)}>
                  {t('layoutTemplates.gridEditor.removeTile')}
                </Button>
              </div>
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
