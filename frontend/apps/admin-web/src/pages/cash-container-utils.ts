import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import type { TFunction } from 'i18next';

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

export function containerOptionLabel(
  container: ListCashBalances200BalancesItem,
  t: TFunction,
): string {
  const containerType = t(`safe.containerTypes.${container.containerType}`, {
    defaultValue: container.containerType,
  });
  return t('safe.containerLabel', {
    containerId: container.containerId,
    containerType,
  });
}
