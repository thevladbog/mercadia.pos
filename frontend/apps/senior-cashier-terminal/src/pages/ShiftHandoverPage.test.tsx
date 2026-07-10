import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, screen, fireEvent, waitFor } from '@testing-library/react';

import { renderWithProviders } from '@/test-utils.js';

import { ShiftHandoverPage } from './ShiftHandoverPage.js';

const CLOSING_SENIOR_ACTOR_ID = 'senior-1';
const SUCCESSOR_ACTOR_ID = 'senior-2';

const navigateMock = vi.fn();
const loginMock = vi.fn();

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

vi.mock('@/auth/AuthProvider.js', () => ({
  useAuth: () => ({
    session: {
      actorId: CLOSING_SENIOR_ACTOR_ID,
      roles: ['senior_cashier'],
      token: 'primary-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
    },
    loggedInAt: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
    login: loginMock,
    logout: vi.fn(),
  }),
}));

vi.mock('@/auth/ibutton.js', () => ({
  readStaffCredential: vi.fn(async () => ({
    factor: { kind: 'ibutton', token: 'demo-token', deviceId: 'device-1', commandId: 'cmd-1' },
  })),
}));

vi.mock('@mercadia/api-clients-store-edge', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@mercadia/api-clients-store-edge')>();
  return {
    ...actual,
    useGetCredentialManagement: vi.fn(),
    useListCashMovements: vi.fn(),
    useListCashRecounts: vi.fn(),
    useListOperationJournal: vi.fn(),
    useListStoreChangeFundRequests: vi.fn(),
    useListStoreReturns: vi.fn(),
  };
});

import {
  useGetCredentialManagement,
  useListCashMovements,
  useListCashRecounts,
  useListOperationJournal,
  useListStoreChangeFundRequests,
  useListStoreReturns,
} from '@mercadia/api-clients-store-edge';

/** Wraps a fixture in the shape `useQuery`-based hooks (from
 * `@mercadia/api-clients-store-edge`) actually return: `{ data: { status,
 * data, headers } }` — the outer `data` is react-query's own field, the
 * inner `status`/`data` is the raw HTTP response `selectSuccessData` reads.
 * Same shape `SafeRecountPage.test.tsx`'s `mockSafeBalance` helper uses. */
function okResponse<T>(data: T) {
  return { data: { status: 200, data, headers: new Headers() } } as unknown;
}

function mockEligibleSuccessors() {
  vi.mocked(useGetCredentialManagement).mockReturnValue(
    okResponse({
      actors: [
        { id: CLOSING_SENIOR_ACTOR_ID, roles: ['senior_cashier'], credentialBindings: [] },
        { id: SUCCESSOR_ACTOR_ID, roles: ['senior_cashier'], credentialBindings: [] },
        { id: 'admin-1', roles: ['admin'], credentialBindings: [] },
        { id: 'cashier-1', roles: ['cashier'], credentialBindings: [] },
      ],
    }) as ReturnType<typeof useGetCredentialManagement>,
  );
}

function mockEmptyOperationalData() {
  vi.mocked(useListCashMovements).mockReturnValue(
    okResponse({ items: [], totalCount: 0 }) as ReturnType<typeof useListCashMovements>,
  );
  vi.mocked(useListStoreReturns).mockReturnValue(
    okResponse({ items: [], totalCount: 0 }) as ReturnType<typeof useListStoreReturns>,
  );
  vi.mocked(useListStoreChangeFundRequests).mockReturnValue(
    okResponse({ items: [], totalCount: 0 }) as ReturnType<typeof useListStoreChangeFundRequests>,
  );
  vi.mocked(useListCashRecounts).mockReturnValue(
    okResponse({ items: [], totalCount: 0 }) as ReturnType<typeof useListCashRecounts>,
  );
  vi.mocked(useListOperationJournal).mockReturnValue(
    okResponse({ items: [] }) as ReturnType<typeof useListOperationJournal>,
  );
}

/** Walks the PIN step (any 4 digits + Enter) then the credential step
 * ("Read credential"), same interaction sequence
 * `SafeRecountPage.test.tsx`'s `SecondSignerAuthModal` flow uses for the
 * shared `PinStep`/`CredentialStep` components. */
async function walkPinAndCredentialSteps() {
  for (const digit of ['1', '2', '3', '4']) {
    fireEvent.click(await screen.findByRole('button', { name: digit }));
  }
  fireEvent.click(screen.getByRole('button', { name: '✓' }));

  fireEvent.click(await screen.findByRole('button', { name: 'Read credential' }));
}

describe('ShiftHandoverPage', () => {
  afterEach(() => {
    cleanup();
  });

  beforeEach(() => {
    navigateMock.mockReset();
    loginMock.mockReset();
    mockEligibleSuccessors();
    mockEmptyOperationalData();
  });

  it('excludes the closing senior cashier from the successor list', async () => {
    renderWithProviders(<ShiftHandoverPage />);

    expect(await screen.findByText(SUCCESSOR_ACTOR_ID)).toBeTruthy();
    expect(screen.getByText('admin-1')).toBeTruthy();
    expect(screen.queryByText(CLOSING_SENIOR_ACTOR_ID)).toBeNull();
    expect(screen.queryByText('cashier-1')).toBeNull();
  });

  it("picks a successor, walks PIN -> credential, and calls login() with the SELECTED SUCCESSOR actorId (not the closing senior's)", async () => {
    loginMock.mockResolvedValue({
      actorId: SUCCESSOR_ACTOR_ID,
      roles: ['senior_cashier'],
      token: 'successor-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
    });

    renderWithProviders(<ShiftHandoverPage />);

    fireEvent.click(await screen.findByText(SUCCESSOR_ACTOR_ID));

    await walkPinAndCredentialSteps();

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledTimes(1);
    });
    expect(loginMock).toHaveBeenCalledWith(
      SUCCESSOR_ACTOR_ID,
      '1234',
      expect.objectContaining({ kind: 'ibutton' }),
    );
    expect(loginMock).not.toHaveBeenCalledWith(
      CLOSING_SENIOR_ACTOR_ID,
      expect.anything(),
      expect.anything(),
    );

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  it("navigates to /monitoring when the successor's real session role is neither senior_cashier nor admin", async () => {
    // Exercises the exact `LoginPage.tsx`-style role branch: navigation is
    // driven by the REAL role(s) `login()` resolves to, not by what the
    // picker showed.
    loginMock.mockResolvedValue({
      actorId: SUCCESSOR_ACTOR_ID,
      roles: ['cashier'],
      token: 'successor-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
    });

    renderWithProviders(<ShiftHandoverPage />);

    fireEvent.click(await screen.findByText(SUCCESSOR_ACTOR_ID));
    await walkPinAndCredentialSteps();

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/monitoring', { replace: true });
    });
  });

  it('shows the carry-over warning banner only when pending counts are non-zero', async () => {
    vi.mocked(useListStoreChangeFundRequests).mockReturnValue(
      okResponse({
        items: [{ id: 'cfr-1', status: 'requested' }],
        totalCount: 1,
      }) as ReturnType<typeof useListStoreChangeFundRequests>,
    );

    renderWithProviders(<ShiftHandoverPage />);

    expect(await screen.findByText('1 change-fund request is still unfulfilled')).toBeTruthy();
  });

  it('hides the carry-over warning banner when there is nothing pending', async () => {
    renderWithProviders(<ShiftHandoverPage />);

    await screen.findByText(SUCCESSOR_ACTOR_ID);
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('changeSuccessor from the PIN step returns to the picker without calling login()', async () => {
    renderWithProviders(<ShiftHandoverPage />);

    fireEvent.click(await screen.findByText(SUCCESSOR_ACTOR_ID));
    fireEvent.click(await screen.findByRole('button', { name: 'Change' }));

    expect(await screen.findByText(SUCCESSOR_ACTOR_ID)).toBeTruthy();
    expect(loginMock).not.toHaveBeenCalled();
  });
});
