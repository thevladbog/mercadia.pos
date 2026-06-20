export type LayoutGridTileSpec = {
  label: string;
  color?: string;
  productId?: string;
  empty?: boolean;
};

export type LayoutGridSpec = {
  rows: number;
  cols: number;
  tiles: LayoutGridTileSpec[];
};
