import { resolveIntlLocale } from '@/i18n/config.js';
import { i18n } from '@/i18n/index.js';

const PAGE_SIZE = 20;

export function defaultReportingWindow(): { since: string; until: string } {
  const until = new Date();
  const since = new Date(until);
  since.setUTCDate(since.getUTCDate() - 7);

  return {
    since: since.toISOString(),
    until: until.toISOString(),
  };
}

export function toDatetimeLocalValue(iso: string): string {
  const date = new Date(iso);
  const pad = (value: number) => String(value).padStart(2, '0');

  return `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())}T${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}`;
}

export function fromDatetimeLocalValue(value: string): string {
  return new Date(`${value}:00.000Z`).toISOString();
}

export function formatMinorAmount(minor: number, locale?: string): string {
  const intlLocale = resolveIntlLocale(locale ?? i18n.language);
  return new Intl.NumberFormat(intlLocale, {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 2,
  }).format(minor / 100);
}

export function formatTimestamp(iso: string, locale?: string): string {
  const intlLocale = resolveIntlLocale(locale ?? i18n.language);
  return new Intl.DateTimeFormat(intlLocale, {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'UTC',
  }).format(new Date(iso));
}

export { PAGE_SIZE };
