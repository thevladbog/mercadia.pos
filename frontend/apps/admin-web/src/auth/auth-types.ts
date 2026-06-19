export type AuthState = {
  userId: string | null;
  roles: string[];
  isAuthenticated: boolean;
};

export type AuthContextValue = AuthState & {
  login: (userId: string, roles: string[], token: string) => void;
  logout: () => void;
  handleUnauthorized: () => void;
};
