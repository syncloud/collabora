export interface FileEntry {
  id: string
  name: string
  size: number
  mtime: string
  kind: string
}

export interface EditorConfig {
  url: string
  token: string
  ttl: number
  name: string
}

export interface SessionInfo {
  username: string
  name: string
  email: string
}

export function login() {
  window.location.href = '/oidc/start'
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { credentials: 'same-origin', ...init })
  if (response.status === 401) {
    login()
    throw new Error('unauthorized')
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error ?? response.statusText)
  }
  return response.json() as Promise<T>
}

export const api = {
  session: () => request<SessionInfo>('/api/session'),
  files: () => request<FileEntry[]>('/api/files'),
  create: (name: string, kind: string) =>
    request<{ id: string; name: string }>('/api/files', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, kind }),
    }),
  upload: (name: string, blob: Blob) =>
    request<{ id: string }>(`/api/files/${encodeURIComponent(fileId(name))}`, {
      method: 'PUT',
      body: blob,
    }),
  remove: (id: string) =>
    request<{ id: string }>(`/api/files/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  editor: (id: string) => request<EditorConfig>(`/api/editor?id=${encodeURIComponent(id)}`),
}

export function fileId(name: string): string {
  const bytes = new TextEncoder().encode(name)
  let binary = ''
  bytes.forEach((byte) => (binary += String.fromCharCode(byte)))
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function humanSize(size: number): string {
  if (size >= 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`
  if (size >= 1024) return `${Math.round(size / 1024)} kB`
  return `${size} B`
}
