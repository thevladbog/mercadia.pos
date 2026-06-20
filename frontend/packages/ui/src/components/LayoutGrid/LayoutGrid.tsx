import type { LayoutGridSpec } from './types.js';
import { cn } from '../../lib/cn.js';

type LayoutGridProps = {
  grid: LayoutGridSpec;
  className?: string;
  onTileClick?: (index: number, tile: LayoutGridSpec['tiles'][number]) => void;
};

export function LayoutGrid({ grid, className, onTileClick }: LayoutGridProps) {
  const slotCount = grid.rows * grid.cols;
  const slots = Array.from(
    { length: slotCount },
    (_, index) => grid.tiles[index] ?? { label: '', empty: true },
  );

  return (
    <div
      className={cn('mercadia-layout-grid', className)}
      style={{ gridTemplateColumns: `repeat(${grid.cols}, minmax(0, 1fr))` }}
    >
      {slots.map((tile, index) => {
        const isEmpty = tile.empty || !tile.label;
        return (
          <button
            key={index}
            className={cn(
              'mercadia-layout-grid-tile',
              isEmpty && 'mercadia-layout-grid-tile--empty',
              !isEmpty && !tile.color && 'mercadia-layout-grid-tile--accent',
            )}
            disabled={isEmpty || !onTileClick}
            onClick={() => onTileClick?.(index, tile)}
            style={tile.color ? { background: tile.color, borderColor: tile.color } : undefined}
            type="button"
          >
            {isEmpty ? '—' : tile.label}
          </button>
        );
      })}
    </div>
  );
}

export type { LayoutGridSpec, LayoutGridTileSpec } from './types.js';
