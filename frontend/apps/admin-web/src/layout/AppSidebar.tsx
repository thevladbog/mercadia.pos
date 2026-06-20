import { useTranslation } from 'react-i18next';
import { NavLink } from 'react-router-dom';

import { useAuth } from '@/auth/useAuth.js';
import { canManageCentralUsers } from '@/auth/permissions.js';

type NavItem = {
  end?: boolean;
  labelKey: string;
  to: string;
};

type NavGroup = {
  items: NavItem[];
  titleKey: string;
};

function SidebarNavGroup({ group }: { group: NavGroup }) {
  const { t } = useTranslation();

  return (
    <div className="sidebar-nav-group">
      <p className="sidebar-nav-group-title">{t(group.titleKey)}</p>
      <ul className="sidebar-nav-list">
        {group.items.map((item) => (
          <li key={item.to}>
            <NavLink
              className={({ isActive }) =>
                `sidebar-nav-link${isActive ? ' sidebar-nav-link--active' : ''}`
              }
              end={item.end}
              to={item.to}
            >
              {t(item.labelKey)}
            </NavLink>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function AppSidebar() {
  const { t } = useTranslation();
  const { roles } = useAuth();
  const isCentralAdmin = canManageCentralUsers(roles);

  const groups: NavGroup[] = [
    {
      titleKey: 'nav.centralGroup',
      items: [
        { to: '/central/dashboard', labelKey: 'nav.dashboard', end: true },
        { to: '/central/reporting', labelKey: 'nav.reporting' },
        { to: '/central/stores', labelKey: 'nav.stores' },
        { to: '/central/sync', labelKey: 'nav.sync' },
        { to: '/central/catalog', labelKey: 'nav.catalog' },
      ],
    },
    {
      titleKey: 'nav.storeGroup',
      items: [
        { to: '/store/monitoring', labelKey: 'nav.monitoring' },
        { to: '/store/safe', labelKey: 'nav.safe' },
        { to: '/store/eod', labelKey: 'nav.eod' },
      ],
    },
  ];

  if (isCentralAdmin) {
    groups.push({
      titleKey: 'nav.adminGroup',
      items: [
        { to: '/central/users', labelKey: 'nav.users' },
        { to: '/central/color-schemes', labelKey: 'nav.colorSchemes' },
        { to: '/central/layout-templates', labelKey: 'nav.layoutTemplates' },
      ],
    });
  }

  return (
    <aside aria-label={t('nav.sidebarLabel')} className="app-sidebar">
      <div className="sidebar-brand">
        <p className="eyebrow">{t('app.eyebrow')}</p>
        <p className="sidebar-brand-title">{t('app.title')}</p>
      </div>
      <nav className="sidebar-nav">
        {groups.map((group) => (
          <SidebarNavGroup key={group.titleKey} group={group} />
        ))}
      </nav>
    </aside>
  );
}
