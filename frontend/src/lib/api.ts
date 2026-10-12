import type { LabBadgeType, LabRequestLog } from '@/types';
import { getStoredAuthToken } from './authContext';
import { getErrorMessage } from './messages';

const MAX_HISTORY_ITEMS = 50;
const HISTORY_STORAGE_KEY = 'labshop_request_history';

// In-memory cache of last 50 logs
let requestHistory: LabRequestLog[] = [];
try {
  const saved = localStorage.getItem(HISTORY_STORAGE_KEY);
  if (saved) {
    requestHistory = JSON.parse(saved);
  }
} catch {
  requestHistory = [];
}

type HistoryListener = (logs: LabRequestLog[]) => void;
const listeners = new Set<HistoryListener>();

function notifyListeners() {
  listeners.forEach((fn) => fn([...requestHistory]));
}

export function getRequestHistory(): LabRequestLog[] {
  return [...requestHistory];
}

export function clearRequestHistory(): void {
  requestHistory = [];
  try {
    localStorage.removeItem(HISTORY_STORAGE_KEY);
  } catch (e) {
    console.warn('Failed to clear request history in localStorage:', e);
  }
  notifyListeners();
}

export function subscribeRequestHistory(listener: HistoryListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function exportHistoryJSON(): string {
  return JSON.stringify(requestHistory, null, 2);
}

function recordLog(log: LabRequestLog) {
  requestHistory.unshift(log);
  if (requestHistory.length > MAX_HISTORY_ITEMS) {
    requestHistory = requestHistory.slice(0, MAX_HISTORY_ITEMS);
  }
  try {
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(requestHistory));
  } catch (e) {
    console.warn('Failed to persist request history to localStorage:', e);
  }
  notifyListeners();
}

export class ApiError extends Error {
  status: number;
  code: string;
  requestId?: string;
  details?: Array<{ field: string; issue: string }>;
  log?: LabRequestLog;
  rawBody?: string;

  constructor(
    message: string,
    status: number,
    code = 'unknown_error',
    requestId?: string,
    details?: Array<{ field: string; issue: string }>,
    log?: LabRequestLog,
    rawBody?: string
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.requestId = requestId;
    this.details = details;
    this.log = log;
    this.rawBody = rawBody;
  }
}

export interface ApiRequestOptions extends RequestInit {
  skipAuth?: boolean;
  idempotencyKey?: string;
}

export interface ApiResponse<T = any> {
  data: T;
  log: LabRequestLog;
  response: Response;
}

/**
 * Single fetch wrapper intercepting all requests to observe origin and edge behaviors.
 */
export async function apiRequest<T = any>(
  inputUrl: string,
  options: ApiRequestOptions = {}
): Promise<ApiResponse<T>> {
  const startTime = performance.now();
  const method = (options.method || 'GET').toUpperCase();

  // Normalize path & URL
  const headers = new Headers(options.headers || {});
  
  // Attach auth token if available & not skipped
  if (!options.skipAuth) {
    const token = getStoredAuthToken();
    if (token && !headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${token}`);
    }
  }

  // Attach idempotency key if provided
  if (options.idempotencyKey && !headers.has('Idempotency-Key')) {
    headers.set('Idempotency-Key', options.idempotencyKey);
  }

  // Set Content-Type default for JSON bodies
  if (options.body && typeof options.body === 'string' && !headers.has('Content-Type')) {
    try {
      JSON.parse(options.body);
      headers.set('Content-Type', 'application/json');
    } catch {
      // not JSON string, let caller handle Content-Type
    }
  }

  const reqHeadersObj: Record<string, string> = {};
  headers.forEach((value, key) => {
    // Redact Authorization token in logs for security
    if (key.toLowerCase() === 'authorization') {
      reqHeadersObj[key] = 'Bearer [REDACTED]';
    } else {
      reqHeadersObj[key] = value;
    }
  });

  const requestBodyPreview = typeof options.body === 'string'
    ? (options.body.length > 500 ? options.body.slice(0, 500) + '...' : options.body)
    : undefined;

  let response: Response;
  try {
    response = await fetch(inputUrl, {
      ...options,
      headers,
    });
  } catch (fetchErr: any) {
    const latencyMs = Math.round(performance.now() - startTime);
    const failedLog: LabRequestLog = {
      id: `req-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      timestamp: new Date().toISOString(),
      method,
      path: inputUrl,
      fullUrl: typeof window !== 'undefined' ? new URL(inputUrl, window.location.href).href : inputUrl,
      status: 0,
      latencyMs,
      requestHeaders: reqHeadersObj,
      responseHeaders: {},
      requestBodyPreview,
      responseBodyPreview: fetchErr?.message || 'Network error / fetch failed',
      badges: ['blocked'],
      speculatedEdgeBlocked: true,
      speculationEvidence: 'Network error occurred. Connection may have been reset or dropped by WAAP/Firewall.',
    };
    recordLog(failedLog);
    throw new ApiError(
      'Không thể kết nối đến máy chủ hoặc yêu cầu đã bị chặn mạng.',
      0,
      'network_error',
      undefined,
      undefined,
      failedLog
    );
  }

  const latencyMs = Math.round(performance.now() - startTime);

  // Extract response headers
  const respHeadersObj: Record<string, string> = {};
  response.headers.forEach((value, key) => {
    respHeadersObj[key.toLowerCase()] = value;
  });

  const originInstance = response.headers.get('x-origin-instance') || undefined;
  const originTime = response.headers.get('x-origin-time') || undefined;
  const requestId = response.headers.get('x-request-id') || undefined;
  const cacheStatus = response.headers.get('x-cache-status') || undefined;
  const age = response.headers.get('age') || undefined;
  const contentEncoding = response.headers.get('content-encoding') || undefined;
  const server = response.headers.get('server') || undefined;
  const via = response.headers.get('via') || undefined;
  const contentType = response.headers.get('content-type') || '';

  // Response classification & Badges according to UIUX_SKILL.md
  const badges: LabBadgeType[] = [];

  if (originInstance) {
    badges.push('origin');
  }

  if (cacheStatus) {
    const cs = cacheStatus.toUpperCase();
    if (cs.includes('HIT')) {
      badges.push('hit');
    } else if (cs.includes('MISS')) {
      badges.push('miss');
    } else if (cs.includes('STALE') || cs.includes('EXPIRED')) {
      badges.push('stale');
    }
  }

  if (contentEncoding && (contentEncoding.includes('gzip') || contentEncoding.includes('br') || contentEncoding.includes('zstd'))) {
    badges.push('compressed');
  }

  // Edge / WAAP block speculation:
  // "Lacks X-Origin-Instance AND (status >= 400 OR non-JSON body) -> Nghi bị chặn ở edge/WAAP"
  const isJson = contentType.includes('application/json');
  const lacksOrigin = !originInstance;
  const isErrorOrNonJson = response.status >= 400 || !isJson;
  const isSpeculatedBlocked = lacksOrigin && isErrorOrNonJson;

  if (isSpeculatedBlocked) {
    badges.push('blocked');
  }

  let evidence = '';
  if (isSpeculatedBlocked) {
    const reasons: string[] = [];
    if (!originInstance) reasons.push('Không có header X-Origin-Instance');
    if (response.status >= 400) reasons.push(`Mã phản hồi HTTP ${response.status}`);
    if (!isJson) reasons.push(`Content-Type không phải JSON (${contentType || 'rỗng'})`);
    evidence = reasons.join('; ');
  }

  // Clone response to read body without consuming stream for caller
  const clone = response.clone();
  let rawBodyText = '';
  let parsedJson: any = null;

  try {
    rawBodyText = await clone.text();
    if (rawBodyText && isJson) {
      parsedJson = JSON.parse(rawBodyText);
    }
  } catch {
    // raw body could not be parsed as text/JSON
  }

  const responseBodyPreview = rawBodyText
    ? (rawBodyText.length > 500 ? rawBodyText.slice(0, 500) + '...' : rawBodyText)
    : undefined;

  const log: LabRequestLog = {
    id: `req-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    timestamp: new Date().toISOString(),
    method,
    path: inputUrl,
    fullUrl: typeof window !== 'undefined' ? new URL(inputUrl, window.location.href).href : inputUrl,
    status: response.status,
    latencyMs,
    requestHeaders: reqHeadersObj,
    responseHeaders: respHeadersObj,
    requestBodyPreview,
    responseBodyPreview,
    badges,
    originInstance,
    originTime,
    requestId,
    cacheStatus,
    age,
    contentEncoding,
    server,
    via,
    speculatedEdgeBlocked: isSpeculatedBlocked,
    speculationEvidence: evidence,
  };

  recordLog(log);

  // If status >= 400, handle structured error
  if (!response.ok) {
    let errorCode = 'http_error';
    let errorMessage = `Yêu cầu thất bại với mã trạng thái ${response.status}`;
    let details: Array<{ field: string; issue: string }> | undefined;

    if (parsedJson && parsedJson.error) {
      errorCode = parsedJson.error.code || errorCode;
      errorMessage = getErrorMessage(errorCode, parsedJson.error.message || errorMessage);
      details = parsedJson.error.details;
    } else if (isSpeculatedBlocked) {
      errorCode = 'edge_blocked';
      errorMessage = getErrorMessage('edge_blocked');
    }

    throw new ApiError(
      errorMessage,
      response.status,
      errorCode,
      requestId,
      details,
      log,
      rawBodyText
    );
  }

  return {
    data: (parsedJson !== null ? parsedJson : (rawBodyText as unknown as T)),
    log,
    response,
  };
}

// Convenience methods
export const api = {
  get: <T = any>(url: string, options?: ApiRequestOptions) =>
    apiRequest<T>(url, { ...options, method: 'GET' }),
  post: <T = any>(url: string, body?: any, options?: ApiRequestOptions) =>
    apiRequest<T>(url, {
      ...options,
      method: 'POST',
      body: body instanceof FormData ? body : (body !== undefined ? JSON.stringify(body) : undefined),
    }),
  put: <T = any>(url: string, body?: any, options?: ApiRequestOptions) =>
    apiRequest<T>(url, {
      ...options,
      method: 'PUT',
      body: body instanceof FormData ? body : (body !== undefined ? JSON.stringify(body) : undefined),
    }),
  patch: <T = any>(url: string, body?: any, options?: ApiRequestOptions) =>
    apiRequest<T>(url, {
      ...options,
      method: 'PATCH',
      body: body instanceof FormData ? body : (body !== undefined ? JSON.stringify(body) : undefined),
    }),
  delete: <T = any>(url: string, options?: ApiRequestOptions) =>
    apiRequest<T>(url, { ...options, method: 'DELETE' }),
};
