export type SyncTab =
  | 'sync-events'
  | 'payments'
  | 'cash-movements'
  | 'fiscal-documents'
  | 'returns'
  | 'operational-days';

export type SyncEntityType = Exclude<SyncTab, 'sync-events'>;

const ENTITY_PATH_SEGMENT: Record<SyncEntityType, string> = {
  payments: 'payments',
  'cash-movements': 'cash-movements',
  'fiscal-documents': 'fiscal-documents',
  returns: 'returns',
  'operational-days': 'operational-days',
};

export const SYNC_ENTITY_PARAM: Record<SyncEntityType, string> = {
  payments: 'paymentId',
  'cash-movements': 'cashMovementId',
  'fiscal-documents': 'fiscalDocumentId',
  returns: 'returnId',
  'operational-days': 'operationalDayId',
};

export const SYNC_ENTITY_LABEL: Record<SyncEntityType, string> = {
  payments: 'Payment',
  'cash-movements': 'Cash movement',
  'fiscal-documents': 'Fiscal document',
  returns: 'Return',
  'operational-days': 'Operational day',
};

export function syncEntityHref(
  storeId: string,
  entityType: SyncEntityType,
  entityId: string,
): string {
  const segment = ENTITY_PATH_SEGMENT[entityType];
  return `/central/sync/stores/${encodeURIComponent(storeId)}/${segment}/${encodeURIComponent(entityId)}`;
}

export function syncExplorerHref(options?: {
  tab?: SyncTab;
  storeId?: string;
  event?: string;
}): string {
  const params = new URLSearchParams();
  if (options?.tab) {
    params.set('tab', options.tab);
  }
  if (options?.storeId) {
    params.set('store', options.storeId);
  }
  if (options?.event) {
    params.set('event', options.event);
  }
  const query = params.toString();
  return query.length > 0 ? `/central/sync?${query}` : '/central/sync';
}

export function readEventFromSearchParams(searchParams: URLSearchParams): string | null {
  return searchParams.get('event');
}

export function parseSyncTab(value: string | null): SyncTab | null {
  const tabs: SyncTab[] = [
    'sync-events',
    'payments',
    'cash-movements',
    'fiscal-documents',
    'returns',
    'operational-days',
  ];
  return tabs.includes(value as SyncTab) ? (value as SyncTab) : null;
}

export function entityTypeFromPathname(pathname: string): SyncEntityType | null {
  for (const [entityType, segment] of Object.entries(ENTITY_PATH_SEGMENT) as [
    SyncEntityType,
    string,
  ][]) {
    if (pathname.includes(`/sync/stores/`) && pathname.includes(`/${segment}/`)) {
      return entityType;
    }
  }
  return null;
}
