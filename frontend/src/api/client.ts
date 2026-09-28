import {Platform} from 'react-native';

import type {ApiErrorBody} from '../types/task';

/**
 * Keputusan #10 — base URL di satu konstanta.
 * Gotcha klasik RN: `localhost` di Android emulator menunjuk emulator itu sendiri,
 * bukan host; `10.0.2.2` adalah alias emulator untuk loopback host.
 * Catatan: perangkat fisik tidak bisa memakai keduanya — ganti ke IP LAN host.
 */
export const BASE_URL: string =
  Platform.select({
    android: 'http://10.0.2.2:8080',
    ios: 'http://localhost:8080',
    default: 'http://localhost:8080',
  }) ?? 'http://localhost:8080';

/** Error bertipe dari API: status HTTP + `code` dari envelope `{error:{code,message}}` BE. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  /** Map field → pesan, hanya ada pada 400 `VALIDATION_ERROR`. */
  readonly fields?: Record<string, string>;

  constructor(
    status: number,
    code: string,
    message: string,
    fields?: Record<string, string>,
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

export interface RequestOptions {
  method?: 'GET' | 'PUT';
  /** Body JSON yang sudah di-stringify. */
  body?: string;
  query?: Record<string, string | number | undefined | null>;
  /** Untuk AbortController di `useTasks` (Fase FE-1): batalkan request basi. */
  signal?: AbortSignal;
}

/** Serialize query params — mengabaikan undefined/null/'' supaya URL tetap bersih. */
export function buildQuery(
  params: RequestOptions['query'] = {},
): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') {
      continue;
    }
    search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `?${qs}` : '';
}

/**
 * Fetch wrapper: JSON in/out, error non-2xx → `ApiError` bertipe.
 * `fetch` TIDAK melempar untuk 4xx/5xx — makanya status dicek manual.
 */
export async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const url = `${BASE_URL}${path}${buildQuery(options.query)}`;

  let response: Response;
  try {
    response = await fetch(url, {
      method: options.method ?? 'GET',
      headers: options.body ? {'Content-Type': 'application/json'} : undefined,
      body: options.body,
      signal: options.signal,
    });
  } catch {
    // Network error (BE mati / salah host) — fetch melempar sebelum dapat Response.
    throw new ApiError(0, 'NETWORK_ERROR', `Tidak bisa menghubungi server di ${BASE_URL}`);
  }

  if (!response.ok) {
    throw await toApiError(response);
  }
  return (await response.json()) as T;
}

/** Parse envelope error BE `{error:{...}}`; fallback kalau body bukan JSON. */
async function toApiError(response: Response): Promise<ApiError> {
  let body: Partial<{error: ApiErrorBody}> | null = null;
  try {
    body = await response.json();
  } catch {
    // Body bukan JSON (mis. HTML dari proxy) — pakai nilai default di bawah.
  }
  const err = body?.error;
  return new ApiError(
    response.status,
    err?.code ?? 'UNKNOWN_ERROR',
    err?.message ?? `Request gagal dengan status ${response.status}`,
    err?.fields,
  );
}
