import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, screen, fireEvent, waitFor } from '@testing-library/react';

import { renderWithProviders } from '@/test-utils.js';

import { SafeRecountPage } from './SafeRecountPage.js';

const PRIMARY_ACTOR_ID = 'senior-1';
const SECOND_SIGNER_ACTOR_ID = 'senior-2';
const SAFE_BALANCE_MINOR = 100_000; // 1,000.00 ₽
const REAL_RECOUNT_ID = 'crec-real-42';

const navigateMock = vi.fn();

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
      actorId: PRIMARY_ACTOR_ID,
      roles: ['senior_cashier'],
      token: 'primary-token',
      expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
    },
    login: vi.fn(),
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
    useListCashBalances: vi.fn(),
    createCashRecount: vi.fn(),
    resolveCashRecount: vi.fn(),
    createAuthSession: vi.fn(),
  };
});

import {
  createAuthSession,
  createCashRecount,
  resolveCashRecount,
  useListCashBalances,
} from '@mercadia/api-clients-store-edge';

function mockSafeBalance(balanceMinor: number) {
  vi.mocked(useListCashBalances).mockReturnValue({
    data: {
      status: 200,
      data: {
        balances: [{ containerId: 'safe-1', containerType: 'safe', balanceMinor }],
      },
      headers: new Headers(),
    },
  } as unknown as ReturnType<typeof useListCashBalances>);
}

/** Types `count` bills of 50 ₽ (5,000 minor units each) into the denomination grid. */
function enterFiftyBillCount(count: string) {
  const input = screen.getByLabelText('50 ₽ — bills');
  fireEvent.change(input, { target: { value: count } });
}

describe('SafeRecountPage', () => {
  afterEach(() => {
    cleanup();
  });

  beforeEach(() => {
    navigateMock.mockReset();
    vi.mocked(createCashRecount).mockReset();
    vi.mocked(resolveCashRecount).mockReset();
    vi.mocked(createAuthSession).mockReset();
    mockSafeBalance(SAFE_BALANCE_MINOR);
  });

  it('creates the recount directly (no approvedById) and navigates to the dashboard when the count matches the safe balance', async () => {
    vi.mocked(createCashRecount).mockResolvedValue({
      status: 202,
      data: {
        recount: {
          id: 'crec-balanced-1',
          storeId: 'store-1',
          containerId: 'safe-1',
          containerType: 'safe',
          currency: 'RUB',
          countedMinor: SAFE_BALANCE_MINOR,
          expectedMinor: SAFE_BALANCE_MINOR,
          discrepancyMinor: 0,
          actorId: PRIMARY_ACTOR_ID,
          status: 'balanced',
          resolutionStatus: 'not_required',
          createdAt: new Date().toISOString(),
        },
      },
      headers: new Headers(),
    } as unknown as Awaited<ReturnType<typeof createCashRecount>>);

    renderWithProviders(<SafeRecountPage />);

    // 20 bills of 50 ₽ == 100,000 minor units == the mocked safe balance.
    enterFiftyBillCount('20');
    fireEvent.click(screen.getByRole('button', { name: 'Confirm' }));

    await waitFor(() => {
      expect(createCashRecount).toHaveBeenCalledTimes(1);
    });

    const [, body] = vi.mocked(createCashRecount).mock.calls[0];
    expect(body).toMatchObject({
      containerType: 'safe',
      containerId: 'safe-1',
      countedMinor: SAFE_BALANCE_MINOR,
      actorId: PRIMARY_ACTOR_ID,
    });
    expect(body).not.toHaveProperty('approvedById');

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
    expect(resolveCashRecount).not.toHaveBeenCalled();
  });

  it('walks a discrepancy through comment -> second-signer auth -> resolution, resolving the REAL created recount id (not a placeholder)', async () => {
    vi.mocked(createCashRecount).mockResolvedValue({
      status: 202,
      data: {
        recount: {
          id: REAL_RECOUNT_ID,
          storeId: 'store-1',
          containerId: 'safe-1',
          containerType: 'safe',
          currency: 'RUB',
          countedMinor: 550_000,
          expectedMinor: SAFE_BALANCE_MINOR,
          discrepancyMinor: 450_000,
          actorId: PRIMARY_ACTOR_ID,
          approvedById: SECOND_SIGNER_ACTOR_ID,
          status: 'discrepancy',
          resolutionStatus: 'open',
          createdAt: '2026-07-10T10:00:00.000Z',
        },
      },
      headers: new Headers(),
    } as unknown as Awaited<ReturnType<typeof createCashRecount>>);

    vi.mocked(createAuthSession).mockResolvedValue({
      status: 201,
      data: {
        session: {
          actorId: SECOND_SIGNER_ACTOR_ID,
          token: 'second-signer-token',
          expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
          roles: ['senior_cashier'],
        },
      },
      headers: new Headers(),
    } as Awaited<ReturnType<typeof createAuthSession>>);

    vi.mocked(resolveCashRecount).mockResolvedValue({
      status: 202,
      data: {
        recount: {
          id: REAL_RECOUNT_ID,
          resolutionStatus: 'resolved',
          resolvedById: SECOND_SIGNER_ACTOR_ID,
        },
      },
      headers: new Headers(),
    } as unknown as Awaited<ReturnType<typeof resolveCashRecount>>);

    renderWithProviders(<SafeRecountPage />);

    // 110 bills of 50 ₽ == 550,000 minor units != the mocked 100,000 safe balance.
    enterFiftyBillCount('110');
    fireEvent.click(screen.getByRole('button', { name: 'Confirm' }));

    // MismatchDialog + the new discrepancy-comment field.
    const resolveButton = (await screen.findByRole('button', {
      name: 'Record discrepancy',
    })) as HTMLButtonElement;
    expect(resolveButton.disabled).toBe(true);

    const commentField = screen.getByPlaceholderText('Describe the reason for the discrepancy…');
    fireEvent.change(commentField, { target: { value: 'Miscounted 90 fifty-ruble notes.' } });
    expect(resolveButton.disabled).toBe(false);
    fireEvent.click(resolveButton);

    // SecondSignerAuthModal — a second, different actor authenticates.
    const idInput = await screen.findByPlaceholderText('Enter ID');
    fireEvent.change(idInput, { target: { value: SECOND_SIGNER_ACTOR_ID } });
    fireEvent.click(screen.getByRole('button', { name: 'Continue' }));

    for (const digit of ['1', '2', '3', '4']) {
      fireEvent.click(await screen.findByRole('button', { name: digit }));
    }
    fireEvent.click(screen.getByRole('button', { name: '✓' }));

    fireEvent.click(await screen.findByRole('button', { name: 'Read credential' }));

    // createCashRecount is only called now — deferred until the second
    // signer is known, so their actorId can satisfy the create-time
    // approval gate for a mismatched count.
    await waitFor(() => {
      expect(createCashRecount).toHaveBeenCalledTimes(1);
    });
    const [, discrepancyBody] = vi.mocked(createCashRecount).mock.calls[0];
    expect(discrepancyBody).toMatchObject({
      containerType: 'safe',
      containerId: 'safe-1',
      countedMinor: 550_000,
      actorId: PRIMARY_ACTOR_ID,
      approvedById: SECOND_SIGNER_ACTOR_ID,
    });

    // RecountResolutionModal — both signatures + summary + the real comment.
    await screen.findByText('Recount Sign-off');
    expect(screen.getByText(PRIMARY_ACTOR_ID)).toBeTruthy();
    expect(screen.getByText(SECOND_SIGNER_ACTOR_ID)).toBeTruthy();
    expect(screen.getByText('Miscounted 90 fifty-ruble notes.')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: '✓ Sign the recount' }));

    await waitFor(() => {
      expect(resolveCashRecount).toHaveBeenCalledTimes(1);
    });

    // The regression assertion for the `recountId` bug: the resolve call's
    // path parameter must be the REAL id from `createCashRecount`'s
    // response, never the old hardcoded `'pending'` placeholder.
    const [resolveStoreId, resolveRecountId, resolveBody] =
      vi.mocked(resolveCashRecount).mock.calls[0];
    expect(resolveStoreId).toBe('store-1');
    expect(resolveRecountId).toBe(REAL_RECOUNT_ID);
    expect(resolveRecountId).not.toBe('pending');
    expect(resolveBody).toMatchObject({
      actorId: SECOND_SIGNER_ACTOR_ID,
      approvedById: PRIMARY_ACTOR_ID,
      resolutionNote: 'Miscounted 90 fifty-ruble notes.',
    });

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });
});
