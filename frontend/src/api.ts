import type { Board, BoardDetail, Card, Column, HistoryEvent, Label, Metrics, Priority, User } from './types'

export class ApiError extends Error {
  status: number
  code?: string
  constructor(message: string, status: number, code?: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

const TOKEN_KEY = 'kanban_token'
export const getToken = (): string | null => {
  try { return localStorage.getItem(TOKEN_KEY) } catch { return null }
}
export const setToken = (t: string | null) => {
  try { t ? localStorage.setItem(TOKEN_KEY, t) : localStorage.removeItem(TOKEN_KEY) } catch { /* sin almacenamiento */ }
}

let onUnauthorized: () => void = () => {}
export const setUnauthorizedHandler = (fn: () => void) => { onUnauthorized = fn }

async function req<T>(method: string, url: string, body?: unknown, publicEndpoint = false): Promise<T> {
  const headers: Record<string, string> = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(url, { method, headers, body: body !== undefined ? JSON.stringify(body) : undefined })
  if (!res.ok) {
    const e = await res.json().catch(() => ({}))
    if (res.status === 401 && !publicEndpoint) onUnauthorized()
    throw new ApiError(e.error || `HTTP ${res.status}`, res.status, e.code)
  }
  return res.status === 204 ? (undefined as T) : res.json()
}

export const api = {
  login: (email: string, password: string) =>
    req<{ token: string; user: User }>('POST', '/api/auth/login', { email, password }, true),
  register: (email: string, name: string, password: string) =>
    req<{ token: string; user: User }>('POST', '/api/auth/register', { email, name, password }, true),
  me: () => req<User>('GET', '/api/me'),
  users: () => req<User[]>('GET', '/api/users'),

  boards: () => req<Board[]>('GET', '/api/boards'),
  board: (id: number) => req<BoardDetail>('GET', `/api/boards/${id}`),
  createBoard: (name: string, description: string) => req<Board>('POST', '/api/boards', { name, description }),
  updateBoard: (id: number, name: string, description: string) =>
    req<void>('PATCH', `/api/boards/${id}`, { name, description }),
  deleteBoard: (id: number) => req<void>('DELETE', `/api/boards/${id}`),
  metrics: (id: number) => req<Metrics>('GET', `/api/boards/${id}/metrics`),

  createColumn: (boardId: number, title: string) => req<Column>('POST', `/api/boards/${boardId}/columns`, { title }),
  updateColumn: (id: number, patch: { title?: string; wip_limit?: number | null; kind?: string }) =>
    req<void>('PATCH', `/api/columns/${id}`, patch),
  deleteColumn: (id: number) => req<void>('DELETE', `/api/columns/${id}`),
  reorderColumns: (boardId: number, ids: number[]) => req<void>('PUT', `/api/boards/${boardId}/columns/order`, { ids }),

  createCard: (column_id: number, title: string, description = '', priority: Priority = 'medium') =>
    req<Card>('POST', '/api/cards', { column_id, title, description, priority }),
  updateCard: (
    id: number,
    patch: { title?: string; description?: string; priority?: Priority; assignee_id?: number | null; label_ids?: number[] },
  ) => req<Card>('PATCH', `/api/cards/${id}`, patch),
  deleteCard: (id: number) => req<void>('DELETE', `/api/cards/${id}`),
  moveCard: (id: number, column_id: number, position: number) =>
    req<void>('PUT', `/api/cards/${id}/move`, { column_id, position }),
  history: (id: number) => req<HistoryEvent[]>('GET', `/api/cards/${id}/history`),

  createLabel: (boardId: number, name: string, color: string) =>
    req<Label>('POST', `/api/boards/${boardId}/labels`, { name, color }),
  updateLabel: (id: number, name: string, color: string) => req<void>('PATCH', `/api/labels/${id}`, { name, color }),
  deleteLabel: (id: number) => req<void>('DELETE', `/api/labels/${id}`),
}
