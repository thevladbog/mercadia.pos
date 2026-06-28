import { useTranslation } from 'react-i18next';
import { Button } from '@mercadia/ui';
import { ThemeToggle } from '@/components/ThemeToggle.js';
import { LanguageSwitcher } from '@/components/LanguageSwitcher.js';

interface TerminalHeaderProps {
  title: string;
  onLogout: () => void;
}

export function TerminalHeader({ title, onLogout }: TerminalHeaderProps) {
  const { t } = useTranslation();

  return (
    <header className="sr-terminal-header">
      <h1>{title}</h1>
      <div className="sr-terminal-header-actions">
        <LanguageSwitcher />
        <ThemeToggle />
        <Button variant="ghost" size="sm" onClick={onLogout}>
          {t('nav.logout')}
        </Button>
      </div>
    </header>
  );
}
