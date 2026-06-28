import { useTranslation } from 'react-i18next';
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { Button } from '@mercadia/ui';
import { changeAppLocale } from '@/i18n/config.js';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const toggle = () => {
    const next = i18n.language === 'ru' ? 'en' : 'ru';
    changeAppLocale(next);
  };

  return (
    <Button variant="ghost" size="sm" onClick={toggle}>
      {i18n.language === 'ru' ? 'EN' : 'RU'}
    </Button>
  );
}
