export function createIdempotencyKey(scope: string, action: string): string {
  return `${scope}:${action}:${crypto.randomUUID()}`;
}

export function createIdempotencyHeaders(idempotencyKey: string): HeadersInit {
  return { 'Idempotency-Key': idempotencyKey };
}
