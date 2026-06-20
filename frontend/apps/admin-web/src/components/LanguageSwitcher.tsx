import { Field, Label, Select } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';

import { changeAppLocale, type AppLocale } from '@/i18n/index.js';
import { SUPPORTED_LOCALES } from '@/i18n/config.js';

const LOCALE_LABEL_KEYS: Record<AppLocale, string> = {
  ru: 'language.ru',
  en: 'language.en',
};

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation();

  return (
    <Field className="language-switcher">
      <Label className="sr-only" htmlFor="admin-language">
        {t('language.label')}
      </Label>
      <Select
        aria-label={t('language.label')}
        id="admin-language"
        value={i18n.language.startsWith('ru') ? 'ru' : 'en'}
        onChange={(event) => changeAppLocale(event.target.value as AppLocale)}
      >
        {SUPPORTED_LOCALES.map((locale) => (
          <option key={locale} value={locale}>
            {t(LOCALE_LABEL_KEYS[locale])}
          </option>
        ))}
      </Select>
    </Field>
  );
}
