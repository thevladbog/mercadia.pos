export type LayoutGridCategorySpec = {
  id: string;
  label: string;
};

export type LayoutGridTileSpec = {
  label: string;
  color?: string;
  productId?: string;
  empty?: boolean;
  categoryId?: string;
  iconUrl?: string;
};

export type LayoutGridSpec = {
  rows: number;
  cols: number;
  categories?: LayoutGridCategorySpec[];
  tiles: LayoutGridTileSpec[];
};
