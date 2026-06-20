import type { LayoutGridSpec } from '@mercadia/ui';

export function defaultGrid(rows = 4, cols = 4): LayoutGridSpec {
  return { rows, cols, categories: [], tiles: [] };
}

export function parseGridJson(raw: string): LayoutGridSpec | null {
  try {
    const parsed = JSON.parse(raw) as LayoutGridSpec;
    if (typeof parsed.rows !== 'number' || typeof parsed.cols !== 'number') {
      return null;
    }
    return gridFromApi(parsed);
  } catch {
    return null;
  }
}

export function gridToApi(grid: LayoutGridSpec) {
  return {
    rows: grid.rows,
    cols: grid.cols,
    ...(grid.categories && grid.categories.length > 0
      ? {
          categories: grid.categories.map((category) => ({
            id: category.id,
            label: category.label,
          })),
        }
      : {}),
    tiles: grid.tiles.map((tile) => ({
      label: tile.label,
      ...(tile.color ? { color: tile.color } : {}),
      ...(tile.productId ? { productId: tile.productId } : {}),
      ...(tile.empty ? { empty: tile.empty } : {}),
      ...(tile.categoryId ? { categoryId: tile.categoryId } : {}),
      ...(tile.iconUrl ? { iconUrl: tile.iconUrl } : {}),
    })),
  };
}

export function gridFromApi(grid: {
  rows?: number;
  cols?: number;
  categories?: { id?: string; label?: string }[];
  tiles?: {
    label?: string;
    color?: string;
    productId?: string;
    empty?: boolean;
    categoryId?: string;
    iconUrl?: string;
  }[];
}): LayoutGridSpec {
  return {
    rows: grid.rows ?? 4,
    cols: grid.cols ?? 4,
    categories: (grid.categories ?? []).map((category) => ({
      id: category.id ?? '',
      label: category.label ?? '',
    })),
    tiles: (grid.tiles ?? []).map((tile) => ({
      label: tile.label ?? '',
      color: tile.color,
      productId: tile.productId,
      empty: tile.empty,
      categoryId: tile.categoryId,
      iconUrl: tile.iconUrl,
    })),
  };
}

export function filterGridByCategory(
  grid: LayoutGridSpec,
  categoryId: string | null,
): LayoutGridSpec {
  if (!categoryId) {
    return grid;
  }
  return {
    ...grid,
    tiles: grid.tiles.filter((tile) => tile.categoryId === categoryId),
  };
}

export function collectLinkedProductIds(grid: LayoutGridSpec): string[] {
  const ids = new Set<string>();
  for (const tile of grid.tiles) {
    if (tile.productId && !tile.empty) {
      ids.add(tile.productId);
    }
  }
  return [...ids];
}

export type PublishValidationResult =
  | { ok: true }
  | { ok: false; reason: 'publishRequiresStore' | 'invalidProducts'; productIds?: string[] };

export function validateGridForPublish(
  grid: LayoutGridSpec,
  storeId: string,
  knownProductIds: ReadonlySet<string>,
): PublishValidationResult {
  if (!storeId) {
    return { ok: false, reason: 'publishRequiresStore' };
  }
  const missing = collectLinkedProductIds(grid).filter(
    (productId) => !knownProductIds.has(productId),
  );
  if (missing.length > 0) {
    return { ok: false, reason: 'invalidProducts', productIds: missing };
  }
  return { ok: true };
}
