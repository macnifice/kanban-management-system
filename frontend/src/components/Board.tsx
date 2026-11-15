import { useState } from 'react'
import {
  DndContext, DragOverlay, PointerSensor, closestCorners, useSensor, useSensors,
  type DragEndEvent, type DragOverEvent, type DragStartEvent,
} from '@dnd-kit/core'
import { SortableContext, arrayMove, horizontalListSortingStrategy } from '@dnd-kit/sortable'
import type { Column as ColumnT, Card as CardT } from '../types'
import { api, ApiError } from '../api'
import { useToast } from './Toast'
import Column from './Column'
import AddForm from './AddForm'
import CardModal from './CardModal'
import ColumnSettings from './ColumnSettings'

interface Props {
  boardId: number
  columns: ColumnT[]
  setColumns: React.Dispatch<React.SetStateAction<ColumnT[]>>
  reload: () => void
}

const cardId = (id: unknown) => Number(String(id).replace('card-', ''))
const colId = (id: unknown) => Number(String(id).replace(/^(col|drop)-/, ''))
const isFull = (c: ColumnT) => c.wip_limit !== null && c.cards.length >= c.wip_limit

export default function Board({ boardId, columns, setColumns, reload }: Props) {
  const toast = useToast()
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }))
  const [active, setActive] = useState<CardT | null>(null)
  const [editing, setEditing] = useState<CardT | null>(null)
  const [settings, setSettings] = useState<ColumnT | null>(null)

  // Ejecuta una mutación, avisa si falla (p. ej. 409 por WIP) y resincroniza con el servidor.
  const safe = async (fn: () => Promise<unknown>) => {
    try {
      await fn()
    } catch (e) {
      const err = e as ApiError
      toast(err.code === 'wip_limit' ? 'La columna destino alcanzó su límite de WIP' : err.message)
    }
    reload()
  }

  const findColumnOf = (id: string) => {
    if (id.startsWith('card-')) return columns.find((c) => c.cards.some((k) => `card-${k.id}` === id))
    return columns.find((c) => c.id === colId(id))
  }

  const onDragStart = (e: DragStartEvent) => {
    if (e.active.data.current?.type === 'card') setActive(e.active.data.current.card)
  }

  // Mueve la tarjeta entre columnas mientras se arrastra (solo estado local).
  // Si la columna destino ya está llena, no se mueve: la regla de WIP se aplica también en el servidor.
  const onDragOver = (e: DragOverEvent) => {
    const { active: a, over } = e
    if (!over || a.data.current?.type !== 'card') return
    const from = findColumnOf(String(a.id))
    const to = findColumnOf(String(over.id))
    if (!from || !to || from.id === to.id) return
    if (isFull(to)) return
    setColumns((cols) => {
      const card = from.cards.find((k) => `card-${k.id}` === a.id)!
      const overIdx = to.cards.findIndex((k) => `card-${k.id}` === over.id)
      const idx = overIdx >= 0 ? overIdx : to.cards.length
      return cols.map((c) => {
        if (c.id === from.id) return { ...c, cards: c.cards.filter((k) => k.id !== card.id) }
        if (c.id === to.id) {
          const cards = [...c.cards]
          cards.splice(idx, 0, { ...card, column_id: to.id })
          return { ...c, cards }
        }
        return c
      })
    })
  }

  const onDragEnd = (e: DragEndEvent) => {
    setActive(null)
    const { active: a, over } = e
    if (!over) return reload()

    if (a.data.current?.type === 'column') {
      const oldI = columns.findIndex((c) => `col-${c.id}` === a.id)
      const newI = columns.findIndex((c) => `col-${c.id}` === over.id)
      if (oldI < 0 || newI < 0 || oldI === newI) return
      const next = arrayMove(columns, oldI, newI)
      setColumns(next)
      safe(() => api.reorderColumns(boardId, next.map((c) => c.id)))
      return
    }

    const col = findColumnOf(String(a.id))
    if (!col) return reload()

    // Soltó sobre una columna distinta que está llena: no se movió, avisamos.
    const target = findColumnOf(String(over.id))
    if (target && target.id !== col.id && isFull(target)) {
      toast(`«${target.title}» alcanzó su límite de WIP (${target.wip_limit})`)
      return reload()
    }

    let cards = col.cards
    const oldI = cards.findIndex((k) => `card-${k.id}` === a.id)
    const overI = cards.findIndex((k) => `card-${k.id}` === over.id)
    if (overI >= 0 && overI !== oldI) {
      cards = arrayMove(cards, oldI, overI)
      setColumns((cs) => cs.map((c) => (c.id === col.id ? { ...c, cards } : c)))
    }
    const position = cards.findIndex((k) => `card-${k.id}` === a.id)
    safe(() => api.moveCard(cardId(a.id), col.id, position))
  }

  return (
    <>
      <DndContext sensors={sensors} collisionDetection={closestCorners}
        onDragStart={onDragStart} onDragOver={onDragOver} onDragEnd={onDragEnd}>
        <div className="board">
          <SortableContext items={columns.map((c) => `col-${c.id}`)} strategy={horizontalListSortingStrategy}>
            {columns.map((c) => (
              <Column key={c.id} column={c}
                onSettings={() => setSettings(c)}
                onDelete={() => safe(() => api.deleteColumn(c.id))}
                onAddCard={(t, d, p) => safe(() => api.createCard(c.id, t, d, p))}
                onOpenCard={setEditing} />
            ))}
          </SortableContext>
          <div className="column new-column">
            <AddForm placeholder="Añadir columna" onAdd={(t) => safe(() => api.createColumn(boardId, t))} />
          </div>
        </div>
        <DragOverlay>{active && <div className="card overlay-card"><div className="card-title">{active.title}</div></div>}</DragOverlay>
      </DndContext>
      {editing && (
        <CardModal card={editing} onClose={() => setEditing(null)}
          onSave={(patch) => safe(() => api.updateCard(editing.id, patch))}
          onDelete={() => safe(() => api.deleteCard(editing.id))} />
      )}
      {settings && (
        <ColumnSettings column={settings} onClose={() => setSettings(null)}
          onSave={(patch) => safe(() => api.updateColumn(settings.id, patch))} />
      )}
    </>
  )
}
