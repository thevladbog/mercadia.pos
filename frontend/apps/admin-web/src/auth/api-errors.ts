import { ApiError } from '@mercadia/api-clients-central';
import { ApiError as StoreEdgeApiError } from '@mercadia/api-clients-store-edge';

export function isUnauthorizedError(error: unknown): boolean {
  return (error instanceof ApiError || error instanceof StoreEdgeApiError) && error.status === 401;
}

export function getApiErrorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof StoreEdgeApiError) {
    return error.problem.detail ?? error.problem.title;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Unexpected error';
}
