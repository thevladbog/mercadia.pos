import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';

export function containersByType(
  balances: ListCashBalances200BalancesItem[],
  containerType: string,
): ListCashBalances200BalancesItem[] {
  return balances.filter((balance) => balance.containerType === containerType);
}

export function firstContainerByType(
  balances: ListCashBalances200BalancesItem[],
  containerType: string,
): ListCashBalances200BalancesItem | undefined {
  return balances.find((balance) => balance.containerType === containerType);
}

export function containerOptionLabel(container: ListCashBalances200BalancesItem): string {
  return `${container.containerId} (${container.containerType})`;
}
