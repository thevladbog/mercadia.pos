import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';

export type SafeBalanceRollups = {
  safeTotalMinor: number;
  drawerTotalMinor: number;
  bankTotalMinor: number;
  nonZeroDrawerCount: number;
  containerCount: number;
};

export function computeSafeBalanceRollups(
  balances: ListCashBalances200BalancesItem[],
): SafeBalanceRollups {
  let safeTotalMinor = 0;
  let drawerTotalMinor = 0;
  let bankTotalMinor = 0;
  let nonZeroDrawerCount = 0;

  for (const balance of balances) {
    switch (balance.containerType) {
      case 'safe':
        safeTotalMinor += balance.balanceMinor;
        break;
      case 'drawer':
        drawerTotalMinor += balance.balanceMinor;
        if (balance.balanceMinor !== 0) {
          nonZeroDrawerCount += 1;
        }
        break;
      case 'bank':
        bankTotalMinor += balance.balanceMinor;
        break;
      default:
        break;
    }
  }

  return {
    safeTotalMinor,
    drawerTotalMinor,
    bankTotalMinor,
    nonZeroDrawerCount,
    containerCount: balances.length,
  };
}

export function countOpenRecountDiscrepancies(items: { resolutionStatus: string }[]): number {
  return items.filter((item) => item.resolutionStatus === 'open').length;
}
