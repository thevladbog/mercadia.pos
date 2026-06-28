import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { configureStoreEdgeClient } from '@/api-client-config.js';
import { Root } from '@/Root.js';
import '@mercadia/ui/styles.css';
import './index.css';

configureStoreEdgeClient();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
);
