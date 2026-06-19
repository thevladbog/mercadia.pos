import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import {
  configureCentralApiClient,
  configureStoreEdgeApiClient,
} from '@/auth/api-client-config.js';
import { Root } from '@/Root.js';
import './index.css';

configureCentralApiClient();
configureStoreEdgeApiClient();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Root />
  </StrictMode>,
);
