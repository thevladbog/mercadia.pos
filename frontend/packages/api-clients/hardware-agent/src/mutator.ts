export class ApiError extends Error {
  readonly problem: {
    code: string;
    detail?: string;
    status: number;
    title: string;
    type: string;
  };

  readonly status: number;

  constructor(status: number, problem: ApiError['problem']) {
    super(problem.title);
    this.name = 'ApiError';
    this.status = status;
    this.problem = problem;
  }
}

let apiBaseUrl = '';

export function setApiBaseUrl(url: string): void {
  const normalized = url.trim().replace(/\/$/, '');
  if (!normalized) {
    throw new Error('VITE_HARDWARE_AGENT_URL is required');
  }
  apiBaseUrl = normalized;
}

export function getApiBaseUrl(): string {
  if (!apiBaseUrl) {
    throw new Error('VITE_HARDWARE_AGENT_URL is required');
  }
  return apiBaseUrl;
}

type FetchEnvelope<TData> = {
  data: TData;
  status: number;
  headers: Headers;
};

export async function customFetch<T extends FetchEnvelope<unknown>>(
  url: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);

  if (options.body !== undefined && options.body !== null && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`${getApiBaseUrl()}${url}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let problem: ApiError['problem'] = {
      type: 'about:blank',
      title: response.statusText || 'Request failed',
      status: response.status,
      code: 'request_failed',
    };

    try {
      const body = (await response.json()) as Partial<ApiError['problem']>;
      if (body.title && body.status && body.code && body.type) {
        problem = {
          type: body.type,
          title: body.title,
          status: body.status,
          code: body.code,
          detail: body.detail,
        };
      }
    } catch {
      // Keep default problem payload.
    }

    throw new ApiError(response.status, problem);
  }

  if (response.status === 204) {
    return {
      data: undefined,
      status: response.status,
      headers: response.headers,
    } as T;
  }

  const data = (await response.json()) as T['data'];

  return {
    data,
    status: response.status,
    headers: response.headers,
  } as T;
}
