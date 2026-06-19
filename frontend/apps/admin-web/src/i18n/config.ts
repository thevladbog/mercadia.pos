export const LOCALE_STORAGE_KEY = 'mercadia.admin.locale';
export const DEFAULT_LOCALE = 'ru';
export const FALLBACK_LOCALE = 'en';
export const SUPPORTED_LOCALES = ['ru', 'en'] as const;

export type AppLocale = (typeof SUPPORTED_LOCALES)[number];

export function resolveIntlLocale(language: string): string {
  if (language.startsWith('ru')) {
    return 'ru-RU';
  }
  return 'en-US';
}

export function readStoredLocale(): AppLocale | null {
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
  if (stored === 'ru' || stored === 'en') {
    return stored;
  }
  return null;
}

export function persistLocale(locale: AppLocale): void {
  localStorage.setItem(LOCALE_STORAGE_KEY, locale);
}
