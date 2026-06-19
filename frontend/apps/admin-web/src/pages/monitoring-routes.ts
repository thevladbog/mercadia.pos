export function terminalMonitoringHref(storeId: string, terminalId: string): string {
  return `/store/monitoring/stores/${encodeURIComponent(storeId)}/terminals/${encodeURIComponent(terminalId)}`;
}

export function monitoringExplorerHref(storeId?: string): string {
  const params = storeId ? `?store=${encodeURIComponent(storeId)}` : '';
  return `/store/monitoring${params}`;
}

export function terminalEventsStreamUrl(storeId: string): string {
  const base = import.meta.env.VITE_STORE_EDGE_URL ?? '';
  return `${base}/v1/stores/${encodeURIComponent(storeId)}/terminals/events`;
}
