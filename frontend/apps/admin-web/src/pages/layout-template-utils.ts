import type { LayoutGridSpec } from '@mercadia/ui';

export function defaultGrid(rows = 4, cols = 4): LayoutGridSpec {
  return { rows, cols, tiles: [] };
}

export function parseGridJson(raw: string): LayoutGridSpec | null {
  try {
    const parsed = JSON.parse(raw) as LayoutGridSpec;
    if (typeof parsed.rows !== 'number' || typeof parsed.cols !== 'number') {
      return null;
    }
    return {
      rows: parsed.rows,
      cols: parsed.cols,
      tiles: parsed.tiles ?? [],
    };
  } catch {
    return null;
  }
}

export function gridToApi(grid: LayoutGridSpec) {
  return {
    rows: grid.rows,
    cols: grid.cols,
    tiles: grid.tiles.map((tile) => ({
      label: tile.label,
      ...(tile.color ? { color: tile.color } : {}),
      ...(tile.productId ? { productId: tile.productId } : {}),
      ...(tile.empty ? { empty: tile.empty } : {}),
    })),
  };
}

export function gridFromApi(grid: {
  rows?: number;
  cols?: number;
  tiles?: { label?: string; color?: string; productId?: string; empty?: boolean }[];
}): LayoutGridSpec {
  return {
    rows: grid.rows ?? 4,
    cols: grid.cols ?? 4,
    tiles: (grid.tiles ?? []).map((tile) => ({
      label: tile.label ?? '',
      color: tile.color,
      productId: tile.productId,
      empty: tile.empty,
    })),
  };
}
