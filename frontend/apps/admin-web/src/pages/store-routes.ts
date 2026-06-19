export function storePageHref(path: string, storeId?: string): string {
  if (!storeId) {
    return path;
  }
  const separator = path.includes('?') ? '&' : '?';
  return `${path}${separator}store=${encodeURIComponent(storeId)}`;
}

export function readStoreFromSearchParams(searchParams: URLSearchParams): string | null {
  return searchParams.get('store');
}
