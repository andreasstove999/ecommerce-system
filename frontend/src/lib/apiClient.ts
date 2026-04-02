import { env } from './env';
import { AppError } from '../utils/errors';
import type { ApiErrorPayload } from '../types/api';

interface RequestOptions extends RequestInit {
  body?: unknown;
}

const buildUrl = (path: string): string => {
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  return `${env.apiBaseUrl}${cleanPath}`;
};

const getErrorMessage = (payload: ApiErrorPayload | null, fallback: string): string => {
  if (payload?.message && payload.message.trim().length > 0) {
    return payload.message;
  }
  return fallback;
};

export const apiRequest = async <T>(path: string, options: RequestOptions = {}): Promise<T> => {
  const { body, headers, ...rest } = options;

  const response = await fetch(buildUrl(path), {
    ...rest,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    let payload: ApiErrorPayload | null = null;

    try {
      payload = (await response.json()) as ApiErrorPayload;
    } catch {
      // noop - fallback to status text below
    }

    throw new AppError(
      getErrorMessage(payload, response.statusText || 'Request failed'),
      response.status,
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
};
