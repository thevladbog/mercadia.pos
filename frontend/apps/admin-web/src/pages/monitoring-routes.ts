export function terminalMonitoringHref(storeId: string, terminalId: string): string {
  return `/store/monitoring/stores/${encodeURIComponent(storeId)}/terminals/${encodeURIComponent(terminalId)}`;
}

export function monitoringExplorerHref(storeId?: string): string {
  const params = storeId ? `?store=${encodeURIComponent(storeId)}` : '';
  return `/store/monitoring${params}`;
}
