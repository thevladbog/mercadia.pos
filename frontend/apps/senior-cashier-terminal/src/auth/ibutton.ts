import {
  listDevices,
  sendDeviceCommand,
  type ListDevices200Item,
} from '@mercadia/api-clients-hardware-agent';

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

  const command = response.data.command;

  if (command.status === 'failed') {
    throw new Error(command.error ?? 'iButton read failed');
  }

  if (typeof command.result?.romId === 'string') {
    return command.result.romId;
  }

  throw new Error('iButton returned no ROM ID');
}
