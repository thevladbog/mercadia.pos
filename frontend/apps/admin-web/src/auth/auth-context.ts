import { createContext } from 'react';

import type { AuthContextValue } from './auth-types.js';

export const AuthContext = createContext<AuthContextValue | null>(null);
