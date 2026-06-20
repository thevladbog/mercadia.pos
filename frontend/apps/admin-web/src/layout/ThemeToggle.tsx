import { Button, useTheme } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';

export function ThemeToggle() {
  const { t } = useTranslation();
  const { theme, setTheme } = useTheme();
  const isDark = theme.colorMode === 'dark';

  return (
    <Button
      aria-label={t('theme.toggle')}
      aria-pressed={isDark}
      onClick={() =>
        setTheme({
          ...theme,
          colorMode: isDark ? 'light' : 'dark',
        })
      }
      type="button"
      variant="secondary"
    >
      {isDark ? t('theme.light') : t('theme.dark')}
    </Button>
  );
}
