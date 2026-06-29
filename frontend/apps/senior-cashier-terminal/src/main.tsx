import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { configureApiClients } from '@/api-client-config.js';
import { Root } from '@/Root.js';
import '@mercadia/ui/styles.css';
import './index.css';

configureApiClients();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
);
