import {
  getDeviceCommand,
  listDevices,
  sendDeviceCommand,
  type GetDeviceCommand200,
  type ListDevices200Item,
  type SendDeviceCommand202Command,
} from '@mercadia/api-clients-hardware-agent';

const COMMAND_POLL_INTERVAL_MS = 50;
const COMMAND_POLL_TIMEOUT_MS = 3_000;

type DeviceCommand = SendDeviceCommand202Command | GetDeviceCommand200;

async function findIButtonDevice(signal?: AbortSignal): Promise<ListDevices200Item | null> {
  try {
    const response = await listDevices({ signal });
    return (
      response.data.find(
        (device) =>
          device.kind === 'ibutton' && (device.status === 'ready' || device.status === 'simulated'),
      ) ?? null
    );
  } catch {
    return null;
  }
}

function wait(ms: number, signal?: AbortSignal): Promise<void> {
  if (signal?.aborted) {
    return Promise.reject(new Error('iButton read aborted'));
  }

  return new Promise((resolve, reject) => {
    const timeoutId = setTimeout(() => {
      signal?.removeEventListener('abort', handleAbort);
      resolve();
    }, ms);
    const handleAbort = () => {
      clearTimeout(timeoutId);
      reject(new Error('iButton read aborted'));
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
  const deadline = Date.now() + COMMAND_POLL_TIMEOUT_MS;

  while (command.status !== 'completed' && command.status !== 'failed') {
    if (Date.now() >= deadline) {
      throw new Error('iButton read timed out');
    }

    await wait(COMMAND_POLL_INTERVAL_MS, signal);
    const response = await getDeviceCommand(deviceId, command.id, { signal });
    if (response.status !== 200) {
      throw new Error('Failed to get iButton command status');
    }
    command = response.data;
  }

  return command;
}

export async function readIButton(signal?: AbortSignal): Promise<string> {
  const device = await findIButtonDevice(signal);
  if (!device) {
    throw new Error('No iButton reader is available');
  }

  const response = await sendDeviceCommand(
    device.id,
    { type: 'read_key' },
    {
      headers: { 'Idempotency-Key': crypto.randomUUID() },
      signal,
    },
  );

  if (response.status !== 202) {
    throw new Error('Failed to send iButton command');
  }

  const command = await waitForCommandCompletion(device.id, response.data.command, signal);

  if (command.status === 'failed') {
    throw new Error(command.error ?? 'iButton read failed');
  }

  if (typeof command.result?.romId === 'string') {
    return command.result.romId;
  }

  throw new Error('iButton returned no ROM ID');
}
