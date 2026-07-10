import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { cleanup, screen, fireEvent, waitFor } from '@testing-library/react';

import { renderWithProviders } from '@/test-utils.js';

import { SecondSignerAuthModal } from './SecondSignerAuthModal.js';

const PRIMARY_ACTOR_ID = 'senior-1';

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
    createAuthSession: vi.fn(),
  };
});

import { createAuthSession } from '@mercadia/api-clients-store-edge';

/** Drives the wizard from step 1 through to the credential-read button click. */
async function fillPersonnelIdAndPin(personnelId: string) {
  const idInput = await screen.findByPlaceholderText('Enter ID');
  fireEvent.change(idInput, { target: { value: personnelId } });
  fireEvent.click(screen.getByRole('button', { name: 'Continue' }));

  for (const digit of ['1', '2', '3', '4']) {
    fireEvent.click(await screen.findByRole('button', { name: digit }));
  }
  fireEvent.click(screen.getByRole('button', { name: '✓' }));

  fireEvent.click(await screen.findByRole('button', { name: 'Read credential' }));
}

describe('SecondSignerAuthModal', () => {
  afterEach(() => {
    cleanup();
  });

  beforeEach(() => {
    vi.mocked(createAuthSession).mockReset();
  });

  it('blocks and shows an inline error when the authenticated actor is the same as the primary session actor', async () => {
    vi.mocked(createAuthSession).mockResolvedValue({
      status: 201,
      data: {
        session: {
          actorId: PRIMARY_ACTOR_ID,
          token: 'same-actor-token',
          expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
          roles: ['senior_cashier'],
        },
      },
      headers: new Headers(),
    } as Awaited<ReturnType<typeof createAuthSession>>);

    const onAuthenticated = vi.fn();
    renderWithProviders(
      <SecondSignerAuthModal open onClose={vi.fn()} onAuthenticated={onAuthenticated} />,
    );

    await fillPersonnelIdAndPin(PRIMARY_ACTOR_ID);

    await waitFor(() => {
      expect(
        screen.getByText(
          'The second signer must be a different person than the current cashier. Choose someone else.',
        ),
      ).toBeTruthy();
    });
    expect(onAuthenticated).not.toHaveBeenCalled();
  });

  it('calls onAuthenticated with a different actor without touching the primary session', async () => {
    vi.mocked(createAuthSession).mockResolvedValue({
      status: 201,
      data: {
        session: {
          actorId: 'senior-2',
          token: 'second-signer-token',
          expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
          roles: ['senior_cashier'],
        },
      },
      headers: new Headers(),
    } as Awaited<ReturnType<typeof createAuthSession>>);

    const onAuthenticated = vi.fn();
    renderWithProviders(
      <SecondSignerAuthModal open onClose={vi.fn()} onAuthenticated={onAuthenticated} />,
    );

    await fillPersonnelIdAndPin('senior-2');

    await waitFor(() => {
      expect(onAuthenticated).toHaveBeenCalledWith('senior-2');
    });
  });
});
