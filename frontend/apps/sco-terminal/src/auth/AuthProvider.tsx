import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import {
  ApiError,
  clearSessionToken,
  createAuthSession,
  setSessionToken,
} from '@mercadia/api-clients-store-edge';

import { envValue, getStoreId, getTerminalId } from '@/api-client-config.js';

import type { SessionResult } from './types.js';

// The SCO terminal has no customer login: a single long-lived service actor authenticates at
// boot and its actorId becomes every customer receipt's cashierId for the terminal's operating
// lifetime (docs/sco-terminal-implementation-design.md §4, Proposal A).
// TODO(sco-auth): placeholder service actor, see docs/sco-terminal-implementation-design.md §4
const SERVICE_ACTOR_ID = envValue('VITE_SCO_SERVICE_ACTOR_ID', 'cashier-1');
const SERVICE_ACTOR_PIN = envValue('VITE_SCO_SERVICE_ACTOR_PIN', '1234');

const SESSION_KEY = 'mercadia.sco-terminal.session';

export type BootState =
  | { status: 'booting' }
  | { status: 'ready'; session: SessionResult }
  | { status: 'error' };

interface AuthContextValue {
  state: BootState;
  retry: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function isSessionResult(value: unknown): value is SessionResult {
  if (typeof value !== 'object' || value === null) return false;
  const candidate = value as Partial<SessionResult>;
  return (
    typeof candidate.token === 'string' &&
    typeof candidate.actorId === 'string' &&
    typeof candidate.expiresAt === 'string' &&
    Array.isArray(candidate.roles)
  );
}

function clearCachedSession(): void {
  sessionStorage.removeItem(SESSION_KEY);
  clearSessionToken();
}

function loadCachedSession(): SessionResult | null {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as unknown;
    if (!isSessionResult(parsed)) {
      clearCachedSession();
      return null;
    }
    const expiresAt = new Date(parsed.expiresAt).getTime();
    if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) {
      clearCachedSession();
      return null;
    }
    setSessionToken(parsed.token);
    return parsed;
  } catch {
    clearCachedSession();
    return null;
  }
}

function saveCachedSession(session: SessionResult): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
  setSessionToken(session.token);
}

/** Performs the boot-time service-actor login. Has no React state of its own, so it is safe to
 * call from any effect without risking a synchronous setState-in-effect. */
async function performServiceLogin(): Promise<SessionResult> {
  const response = await createAuthSession({
    actorId: SERVICE_ACTOR_ID,
    pin: SERVICE_ACTOR_PIN,
    storeId: getStoreId(),
    terminalId: getTerminalId(),
  });
  if (response.status !== 201) {
    throw new Error('Service actor authentication failed');
  }
  saveCachedSession(response.data.session);
  return response.data.session;
}

function logServiceLoginFailure(error: unknown): void {
  if (error instanceof ApiError) {
    console.error(
      `[sco-terminal] service actor login failed: ${error.status} ${error.problem.code}`,
      error.problem.detail,
    );
  } else {
    console.error('[sco-terminal] service actor login failed', error);
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<BootState>(() => {
    const cached = loadCachedSession();
    return cached ? { status: 'ready', session: cached } : { status: 'booting' };
  });
  const [loginAttempt, setLoginAttempt] = useState(0);
  const refreshTimeoutRef = useRef<number | null>(null);

  // Boot-time (and retry-triggered) login. Defined inline so every setState call happens after
  // an await, inside a function this effect owns directly.
  useEffect(() => {
    if (state.status !== 'booting') {
      return;
    }
    let cancelled = false;
    async function attemptLogin(): Promise<void> {
      try {
        const session = await performServiceLogin();
        if (!cancelled) {
          setState({ status: 'ready', session });
        }
      } catch (error) {
        logServiceLoginFailure(error);
        clearCachedSession();
        if (!cancelled) {
          setState({ status: 'error' });
        }
      }
    }
    void attemptLogin();
    return () => {
      cancelled = true;
    };
  }, [state, loginAttempt]);

  // Silently refresh the service session before it expires. Store Edge does not require a valid
  // session token for the basic checkout operations SCO uses in M1 (cashierId/actorId are passed
  // explicitly in each request body), so a slow refresh here does not interrupt an in-progress
  // customer receipt.
  useEffect(() => {
    if (refreshTimeoutRef.current !== null) {
      window.clearTimeout(refreshTimeoutRef.current);
      refreshTimeoutRef.current = null;
    }
    if (state.status !== 'ready') {
      return;
    }
    const expiresAt = new Date(state.session.expiresAt).getTime();
    if (!Number.isFinite(expiresAt)) {
      return;
    }
    let cancelled = false;
    const delayMs = Math.max(expiresAt - Date.now() - 30_000, 0);
    refreshTimeoutRef.current = window.setTimeout(() => {
      void (async () => {
        try {
          const session = await performServiceLogin();
          if (!cancelled) {
            setState({ status: 'ready', session });
          }
        } catch (error) {
          logServiceLoginFailure(error);
          clearCachedSession();
          if (!cancelled) {
            setState({ status: 'error' });
          }
        }
      })();
    }, delayMs);
    return () => {
      cancelled = true;
      if (refreshTimeoutRef.current !== null) {
        window.clearTimeout(refreshTimeoutRef.current);
      }
    };
  }, [state]);

  const retry = useCallback(() => {
    setState({ status: 'booting' });
    setLoginAttempt((attempt) => attempt + 1);
  }, []);

  const value = useMemo(() => ({ state, retry }), [state, retry]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
