export type StorybookLocale = 'en' | 'ru';

export function getStorybookLocale(value: unknown): StorybookLocale {
  return value === 'ru' ? 'ru' : 'en';
}
