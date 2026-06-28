export interface SessionResult {
  token: string;
  actorId: string;
  roles: string[];
  expiresAt: string;
}

export interface IButtonResult {
  romId: string;
}

export interface HardwareAgentCommandResponse {
  id: string;
  deviceId: string;
  type: string;
  status: string;
  result: IButtonResult | null;
  error: string | null;
}

export interface HardwareAgentDevice {
  id: string;
  kind: string;
  status: string;
  model: string;
}
