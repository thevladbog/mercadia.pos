export type TerminalHeartbeatEvent = {
  type: 'terminal_heartbeat';
  terminalId: string;
  storeId: string;
  kind: string;
  status: string;
  softwareVersion?: string;
  lastSeenAt: string;
  updatedAt: string;
  receivedAt: string;
};

export type TerminalHeartbeatPayload = Omit<TerminalHeartbeatEvent, 'receivedAt'>;
