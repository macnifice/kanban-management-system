export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type Kind = 'normal' | 'active' | 'done'

export interface Label {
  id: number
  board_id: number
  name: string
  color: string
}

export interface Card {
  id: number
  column_id: number
  title: string
  description: string
  position: number
  priority: Priority
  assignee_id: number | null
  label_ids: number[]
  created_at: string
}

export interface Column {
  id: number
  board_id: number
  title: string
  position: number
  wip_limit: number | null
  kind: Kind
  cards: Card[]
}

export interface Board {
  id: number
  name: string
  description: string
  created_at: string
  card_count: number
}

export interface BoardDetail extends Board {
  columns: Column[]
  labels: Label[]
}

export interface User {
  id: number
  email: string
  name: string
}

export interface HistoryEvent {
  id: number
  moved_at: string
  from_column: string | null
  to_column: string
  to_kind: Kind
  user: string | null
}

export interface Metrics {
  columns: { id: number; title: string; kind: Kind; count: number; wip_limit: number | null }[]
  totals_by_kind: Record<Kind, number>
  total_cards: number
  cycle_time: {
    avg_hours: number
    median_hours: number
    count: number
    recent: { card_id: number; title: string; hours: number }[]
  }
}

export const PRIORITIES: { value: Priority; label: string; color: string }[] = [
  { value: 'low', label: 'Baja', color: '#52a37a' },
  { value: 'medium', label: 'Media', color: '#2f6fed' },
  { value: 'high', label: 'Alta', color: '#e08a1e' },
  { value: 'urgent', label: 'Urgente', color: '#d64545' },
]

export const KINDS: { value: Kind; label: string; hint: string }[] = [
  { value: 'normal', label: 'Normal', hint: 'Backlog, pruebas… no cuenta para el cycle time' },
  { value: 'active', label: 'En curso', hint: 'Al entrar aquí empieza el cycle time' },
  { value: 'done', label: 'Terminado', hint: 'Al entrar aquí termina el cycle time' },
]

// Paleta fija para etiquetas (tonos con buen contraste con texto blanco).
export const LABEL_COLORS: { value: string; name: string }[] = [
  { value: '#e53e3e', name: 'Rojo' },
  { value: '#ed8936', name: 'Naranja' },
  { value: '#d69e2e', name: 'Amarillo' },
  { value: '#38a169', name: 'Verde' },
  { value: '#319795', name: 'Turquesa' },
  { value: '#3182ce', name: 'Azul' },
  { value: '#5a67d8', name: 'Índigo' },
  { value: '#805ad5', name: 'Morado' },
  { value: '#d53f8c', name: 'Rosa' },
  { value: '#718096', name: 'Gris' },
]
