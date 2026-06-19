export const MONITORING_REFRESH_INTERVAL_MS = 5000;

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
