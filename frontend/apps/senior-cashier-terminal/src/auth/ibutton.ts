import type { HardwareAgentDevice, HardwareAgentCommandResponse } from './types.js';

const HARDWARE_AGENT_BASE = import.meta.env.VITE_HARDWARE_AGENT_URL ?? '';

async function findIButtonDevice(): Promise<HardwareAgentDevice | null> {
  const res = await fetch(`${HARDWARE_AGENT_BASE}/v1/devices`);
  if (!res.ok) {
    return null;
  }
  const data = (await res.json()) as { devices: HardwareAgentDevice[] };
  return data.devices.find((d) => d.kind === 'ibutton' && d.status === 'ready') ?? null;
}

export async function readIButton(signal?: AbortSignal): Promise<string> {
  let device = await findIButtonDevice();
  if (!device) {
    device = { id: 'ibutton-sim', kind: 'ibutton', status: 'ready', model: 'Simulated iButton' };
  }

  const commandRes = await fetch(`${HARDWARE_AGENT_BASE}/v1/devices/${device.id}/commands`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ type: 'read_key' }),
    signal,
  });

  if (!commandRes.ok) {
    throw new Error('Failed to send iButton command');
  }

  const commandData = (await commandRes.json()) as HardwareAgentCommandResponse;

  if (commandData.status === 'failed') {
    throw new Error(commandData.error ?? 'iButton read failed');
  }

  if (commandData.result?.romId) {
    return commandData.result.romId;
  }

  throw new Error('iButton returned no ROM ID');
}
