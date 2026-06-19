let unauthorizedHandler: (() => void) | null = null;

export function registerUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler;
}

export function unregisterUnauthorizedHandler(): void {
  unauthorizedHandler = null;
}

export function notifyUnauthorized(): void {
  unauthorizedHandler?.();
}
