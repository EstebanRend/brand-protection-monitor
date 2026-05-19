import type { Keyword, MatchedCertificate, MonitorStatus } from '../types/models';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const hasBody = options?.body !== undefined;
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      ...(hasBody ? { 'Content-Type': 'application/json' } : {}),
      ...(options?.headers ?? {}),
    },
    ...options,
  });

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(errorBody.error ?? 'Request failed');
  }

  return response.json() as Promise<T>;
}

export const api = {
  listKeywords: () => request<Keyword[]>('/api/keywords'),
  createKeyword: (value: string) =>
    request<Keyword>('/api/keywords', {
      method: 'POST',
      body: JSON.stringify({ value }),
    }),
  deleteKeyword: (id: number) =>
    fetch(`${API_BASE_URL}/api/keywords/${id}`, { method: 'DELETE' }).then((response) => {
      if (!response.ok) throw new Error('Failed to delete keyword');
    }),
  listMatches: () => request<MatchedCertificate[]>('/api/matches'),
  getStatus: () => request<MonitorStatus>('/api/status'),
  runOnce: () => request<{ status: string }>('/api/monitor/run-once', { method: 'POST' }),
  exportUrl: `${API_BASE_URL}/api/export.csv`,
};
