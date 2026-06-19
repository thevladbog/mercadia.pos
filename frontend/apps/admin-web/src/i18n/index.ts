import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import {
  DEFAULT_LOCALE,
  FALLBACK_LOCALE,
  readStoredLocale,
  type AppLocale,
  persistLocale,
} from './config.js';
import en from './locales/en.json';
import ru from './locales/ru.json';

export function initI18n(): typeof i18n {
  const stored = readStoredLocale();

  void i18n.use(initReactI18next).init({
    resources: {
      en: { translation: en },
      ru: { translation: ru },
    },
    lng: stored ?? DEFAULT_LOCALE,
    fallbackLng: FALLBACK_LOCALE,
    interpolation: { escapeValue: false },
  });

  return i18n;
}

export function changeAppLocale(locale: AppLocale): void {
  persistLocale(locale);
  void i18n.changeLanguage(locale);
}

export { i18n };
export type { AppLocale } from './config.js';
