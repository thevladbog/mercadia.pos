import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from './locales/en.json' with { type: 'json' };
import ru from './locales/ru.json' with { type: 'json' };

const LOCALE_KEY = 'mercadia.sr-terminal.locale';

function getInitialLocale(): string {
  try {
    return localStorage.getItem(LOCALE_KEY) ?? 'ru';
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

export function changeAppLocale(locale: string): void {
  i18n.changeLanguage(locale);
  try {
    localStorage.setItem(LOCALE_KEY, locale);
  } catch {
    /* noop */
  }
}
