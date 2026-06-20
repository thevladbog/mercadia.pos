export function storePageHref(path: string, storeId?: string): string {
  if (!storeId) {
    return path;
  }
  const separator = path.includes('?') ? '&' : '?';
  return `${path}${separator}store=${encodeURIComponent(storeId)}`;
}

export function safeRecountHref(storeId: string, recountId: string): string {
  return `${storePageHref('/store/safe', storeId)}&recount=${encodeURIComponent(recountId)}`;
}

export function safeMovementHref(storeId: string, movementId: string): string {
  return `${storePageHref('/store/safe', storeId)}&movement=${encodeURIComponent(movementId)}`;
}

export function readStoreFromSearchParams(searchParams: URLSearchParams): string | null {
  return searchParams.get('store');
}

export function readRecountFromSearchParams(searchParams: URLSearchParams): string | null {
  return searchParams.get('recount');
}

export function readMovementFromSearchParams(searchParams: URLSearchParams): string | null {
  return searchParams.get('movement');
}
