import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from './locales/en.json' with { type: 'json' };
import ru from './locales/ru.json' with { type: 'json' };

const LOCALE_KEY = 'mercadia.pos-terminal.locale';
const SUPPORTED_LOCALES = ['ru', 'en'] as const;

export type AppLocale = (typeof SUPPORTED_LOCALES)[number];

function normalizeLocale(locale: string | null): AppLocale {
  return SUPPORTED_LOCALES.includes(locale as AppLocale) ? (locale as AppLocale) : 'ru';
}

function getInitialLocale(): AppLocale {
  try {
    return normalizeLocale(localStorage.getItem(LOCALE_KEY));
  } catch {
    return 'ru';
  }
}

export const i18n = i18next.createInstance();

i18n.use(initReactI18next).init({
  lng: getInitialLocale(),
  fallbackLng: 'en',
  resources: {
    ru: { translation: ru },
    en: { translation: en },
  },
  interpolation: {
    escapeValue: false,
  },
});

export function changeAppLocale(locale: AppLocale): void {
  void i18n.changeLanguage(locale);
  try {
    localStorage.setItem(LOCALE_KEY, locale);
  } catch {
    // Ignore persistence failures in restricted browser contexts.
  }
}
