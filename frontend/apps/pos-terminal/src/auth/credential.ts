import {
  getDeviceCommand,
  listDevices,
  sendDeviceCommand,
  type GetDeviceCommand200,
  type ListDevices200Item,
  type SendDeviceCommand202Command,
} from '@mercadia/api-clients-hardware-agent';
import type { CreateAuthSessionBodyCredentialFactor } from '@mercadia/api-clients-store-edge';

const COMMAND_POLL_INTERVAL_MS = 50;
const CREDENTIAL_READ_TIMEOUT_MS = 3_000;

type DeviceCommand = SendDeviceCommand202Command | GetDeviceCommand200;
export type StaffCredentialKind = CreateAuthSessionBodyCredentialFactor['kind'];

export type StaffCredentialRead = {
  factor: CreateAuthSessionBodyCredentialFactor;
  maskedToken?: string;
};

const CREDENTIAL_COMMANDS: Record<
  StaffCredentialKind,
  {
    commandType: string;
    deviceKind: string;
    tokenField: 'romId' | 'staffToken';
  }
> = {
  ibutton: { commandType: 'read_key', deviceKind: 'ibutton', tokenField: 'romId' },
  msr_card: { commandType: 'read_staff_card', deviceKind: 'msr', tokenField: 'staffToken' },
  barcode_card: { commandType: 'scan_staff_card', deviceKind: 'scanner', tokenField: 'staffToken' },
};

async function findCredentialDevice(
  kind: StaffCredentialKind,
  signal?: AbortSignal,
): Promise<ListDevices200Item | null> {
  const config = CREDENTIAL_COMMANDS[kind];
  try {
    const response = await listDevices({ signal });
    return (
      response.data.find(
        (device) =>
          device.kind === config.deviceKind &&
          (device.status === 'ready' || device.status === 'simulated'),
      ) ?? null
    );
  } catch {
    if (signal?.aborted) {
      throw new Error('Staff credential read aborted');
    }
    return null;
  }
}

function wait(ms: number, signal?: AbortSignal): Promise<void> {
  if (signal?.aborted) {
    return Promise.reject(new Error('Staff credential read aborted'));
  }

  return new Promise((resolve, reject) => {
    const timeoutId = setTimeout(() => {
      signal?.removeEventListener('abort', handleAbort);
      resolve();
    }, ms);
    const handleAbort = () => {
      clearTimeout(timeoutId);
      reject(new Error('Staff credential read aborted'));
    };
    signal?.addEventListener('abort', handleAbort, { once: true });
  });
}

async function waitForCommandCompletion(
  deviceId: string,
  initialCommand: DeviceCommand,
  signal?: AbortSignal,
): Promise<DeviceCommand> {
  let command = initialCommand;

  while (command.status !== 'completed' && command.status !== 'failed') {
    await wait(COMMAND_POLL_INTERVAL_MS, signal);
    const response = await getDeviceCommand(deviceId, command.id, { signal });
    if (response.status !== 200) {
      throw new Error('Failed to get staff credential command status');
    }
    command = response.data;
  }

  return command;
}

function createTimeoutSignal(signal?: AbortSignal): { signal: AbortSignal; cleanup: () => void } {
  const abortController = new AbortController();
  const timeoutId = window.setTimeout(() => {
    abortController.abort(new Error('Staff credential read timed out'));
  }, CREDENTIAL_READ_TIMEOUT_MS);
  const handleAbort = () => {
    abortController.abort(new Error('Staff credential read aborted'));
  };

  if (signal?.aborted) {
    handleAbort();
  } else {
    signal?.addEventListener('abort', handleAbort, { once: true });
  }

  return {
    signal: abortController.signal,
    cleanup: () => {
      window.clearTimeout(timeoutId);
      signal?.removeEventListener('abort', handleAbort);
    },
  };
}

export async function readStaffCredential(
  kind: StaffCredentialKind,
  signal?: AbortSignal,
): Promise<StaffCredentialRead> {
  const readSignal = createTimeoutSignal(signal);
  const config = CREDENTIAL_COMMANDS[kind];
  try {
    const device = await findCredentialDevice(kind, readSignal.signal);
    if (!device) {
      throw new Error('No staff credential reader is available');
    }

    const response = await sendDeviceCommand(
      device.id,
      { type: config.commandType },
      {
        headers: { 'Idempotency-Key': crypto.randomUUID() },
        signal: readSignal.signal,
      },
    );

    if (response.status !== 202) {
      throw new Error('Failed to send staff credential command');
    }

    const command = await waitForCommandCompletion(
      device.id,
      response.data.command,
      readSignal.signal,
    );

    if (command.status === 'failed') {
      throw new Error(command.error ?? 'Staff credential read failed');
    }

    const token = command.result?.[config.tokenField];
    if (typeof token !== 'string') {
      throw new Error('Staff credential command returned no token');
    }

    const maskedToken = command.result?.masked;
    return {
      factor: {
        kind,
        token,
        deviceId: device.id,
        commandId: command.id,
      },
      maskedToken: typeof maskedToken === 'string' ? maskedToken : undefined,
    };
  } finally {
    readSignal.cleanup();
  }
}
