import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { configureCentralApiClient } from '@/api-client-config.js';
import { Root } from '@/Root.js';
import '@mercadia/ui/styles.css';
import './index.css';

configureCentralApiClient();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
);
