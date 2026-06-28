import { createContext, useContext, useState, useCallback, useMemo, type ReactNode } from 'react';

import type { SessionResult } from './types.js';
import { readIButton } from './ibutton.js';

export { readIButton };

interface AuthContextValue {
  session: SessionResult | null;
  login: (actorId: string, pin: string, ibuttonRomId: string) => Promise<SessionResult>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const SESSION_KEY = 'mercadia.sr-terminal.session';

function loadSession(): SessionResult | null {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as SessionResult;
    if (new Date(parsed.expiresAt) <= new Date()) {
      sessionStorage.removeItem(SESSION_KEY);
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function saveSession(session: SessionResult): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

function clearSession(): void {
  sessionStorage.removeItem(SESSION_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<SessionResult | null>(loadSession);

  const login = useCallback(async (actorId: string, pin: string, _ibuttonRomId: string): Promise<SessionResult> => {
    const url = import.meta.env.VITE_STORE_EDGE_URL
      ? `${import.meta.env.VITE_STORE_EDGE_URL}/v1/auth/sessions`
      : '/v1/auth/sessions';

    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ actorId, pin }),
    });

    if (!res.ok) {
      if (res.status === 401) {
        throw new Error('Invalid credentials');
      }
      throw new Error('Authentication failed');
    }

    const data = await res.json() as { session: SessionResult };
    saveSession(data.session);
    setSession(data.session);
    return data.session;
  }, []);

  const logout = useCallback(() => {
    clearSession();
    setSession(null);
  }, []);

  const value = useMemo(() => ({ session, login, logout }), [session, login, logout]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
