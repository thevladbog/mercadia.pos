import { useEffect, useState } from 'react';

import { terminalEventsStreamUrl } from './monitoring-routes.js';
import type { StreamConnectionStatus } from './monitoring-utils.js';
import type { TerminalHeartbeatEvent, TerminalHeartbeatPayload } from './terminal-events-types.js';

const DEFAULT_MAX_EVENTS = 50;

type UseTerminalHeartbeatStreamOptions = {
  terminalId?: string;
  maxEvents?: number;
};

function parseHeartbeatEvent(data: string): TerminalHeartbeatPayload | null {
  try {
    const parsed = JSON.parse(data) as TerminalHeartbeatPayload;
    if (parsed.type !== 'terminal_heartbeat' || !parsed.terminalId) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export function useTerminalHeartbeatStream(
  storeId: string,
  options?: UseTerminalHeartbeatStreamOptions,
) {
  const terminalIdFilter = options?.terminalId;
  const maxEvents = options?.maxEvents ?? DEFAULT_MAX_EVENTS;
  const enabled = storeId.length > 0;
  const [events, setEvents] = useState<TerminalHeartbeatEvent[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<StreamConnectionStatus>('connecting');

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const source = new EventSource(terminalEventsStreamUrl(storeId));

    const handleHeartbeat = (message: MessageEvent<string>) => {
      const parsed = parseHeartbeatEvent(message.data);
      if (parsed == null) {
        return;
      }
      if (terminalIdFilter != null && parsed.terminalId !== terminalIdFilter) {
        return;
      }

      const event: TerminalHeartbeatEvent = {
        ...parsed,
        receivedAt: new Date().toISOString(),
      };
      setEvents((current) => [event, ...current].slice(0, maxEvents));
    };

    source.addEventListener('terminal_heartbeat', handleHeartbeat);
    source.onmessage = handleHeartbeat;
    source.onopen = () => setConnectionStatus('connected');
    source.onerror = () => setConnectionStatus('error');

    return () => {
      source.removeEventListener('terminal_heartbeat', handleHeartbeat);
      source.close();
    };
  }, [enabled, storeId, terminalIdFilter, maxEvents]);

  return {
    events: enabled ? events : [],
    connectionStatus: enabled ? connectionStatus : 'idle',
  };
}
