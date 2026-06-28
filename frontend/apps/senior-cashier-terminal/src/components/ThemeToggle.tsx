import { useTheme } from '@mercadia/ui';
import { useTranslation } from 'react-i18next';
import { Button } from '@mercadia/ui';

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const { t } = useTranslation();

  const isDark = theme.colorMode === 'dark';

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => setTheme({ ...theme, colorMode: isDark ? 'light' : 'dark' })}
      title={t('theme.toggle')}
      aria-label={t('theme.toggle')}
    >
      {isDark ? '☀️' : '🌙'}
    </Button>
  );
}
