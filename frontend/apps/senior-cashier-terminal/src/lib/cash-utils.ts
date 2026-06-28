const RUBLE_DENOMINATIONS = [
  { value: 5000_00, label: '5 000' },
  { value: 2000_00, label: '2 000' },
  { value: 1000_00, label: '1 000' },
  { value: 500_00, label: '500' },
  { value: 200_00, label: '200' },
  { value: 100_00, label: '100' },
  { value: 50_00, label: '50' },
  { value: 10_00, label: '10' },
  { value: 5_00, label: '5' },
  { value: 2_00, label: '2' },
  { value: 1_00, label: '1' },
] as const;

interface DenominationValues {
  [denominationMinor: number]: string;
}

export function getDenominations(): readonly { value: number; label: string }[] {
  return RUBLE_DENOMINATIONS;
}

export function computeDenominationTotal(values: DenominationValues, otherAmountMinor: number = 0): number {
  let total = otherAmountMinor;
  for (const [denomStr, countStr] of Object.entries(values)) {
    const denom = Number(denomStr);
    const count = Number(countStr) || 0;
    total += denom * count;
  }
  return total;
}

export function parseRublesToMinor(rubles: string): number {
  const cleaned = rubles.replace(/\s/g, '').replace(',', '.');
  const num = parseFloat(cleaned);
  if (Number.isNaN(num)) return 0;
  return Math.round(num * 100);
}

export function formatMinor(amount: number): string {
  return (amount / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function actorsMustDiffer(actorId: string, approvedById: string): boolean {
  return actorId !== approvedById;
}

export function createIdempotencyHeaders(): Record<string, string> {
  return { 'Idempotency-Key': crypto.randomUUID() };
}
