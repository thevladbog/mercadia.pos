export const MONITORING_REFRESH_INTERVAL_MS = 5000;

export type StreamConnectionStatus = 'idle' | 'connecting' | 'connected' | 'error';

export function streamConnectionStatusLabel(status: StreamConnectionStatus): string {
  switch (status) {
    case 'connecting':
      return 'Connecting';
    case 'connected':
      return 'Connected';
    case 'error':
      return 'Error';
    default:
      return 'Disconnected';
  }
}

export function streamConnectionStatusClass(status: StreamConnectionStatus): string {
  switch (status) {
    case 'connected':
      return 'status-badge status-online';
    case 'connecting':
      return 'status-badge status-attention';
    case 'error':
      return 'status-badge status-offline';
    default:
      return 'status-badge';
  }
}

type TerminalStatusInput = {
  attentionNeeded?: boolean;
  status: string;
};

export function terminalStatusLabel(terminal: TerminalStatusInput): string {
  if (terminal.attentionNeeded) {
    return 'attention';
  }
  return terminal.status;
}

export function terminalStatusClass(terminal: TerminalStatusInput): string {
  if (terminal.attentionNeeded) {
    return 'status-badge status-attention';
  }
  if (terminal.status === 'offline') {
    return 'status-badge status-offline';
  }
  return 'status-badge status-online';
}
