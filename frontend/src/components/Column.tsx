import { SortableContext, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useDroppable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Column as ColumnT, Card as CardT, Priority } from '../types'
import Card from './Card'
import AddForm from './AddForm'

interface Props {
  column: ColumnT
  onSettings: () => void
  onDelete: () => void
  onAddCard: (title: string, description: string, priority: Priority) => void
  onOpenCard: (card: CardT) => void
}

export default function Column({ column, onSettings, onDelete, onAddCard, onOpenCard }: Props) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({
    id: `col-${column.id}`,
    data: { type: 'column' },
  })
  const { setNodeRef: setDropRef } = useDroppable({ id: `drop-${column.id}`, data: { type: 'drop', columnId: column.id } })

  const n = column.cards.length
  const limit = column.wip_limit
  const full = limit !== null && n >= limit
  const over = limit !== null && n > limit
  const kindTag = column.kind === 'active' ? 'en curso' : column.kind === 'done' ? 'terminado' : ''

  return (
    <div ref={setNodeRef} className={`column ${over ? 'col-over' : full ? 'col-full' : ''}`}
      style={{ transform: CSS.Transform.toString(transform), transition }}>
      <div className="column-header" {...attributes} {...listeners}>
        <h3>
          {column.title}
          <span className={`wip ${over ? 'bad' : full ? 'warn' : ''}`} title={limit === null ? 'Sin límite de WIP' : `Límite de WIP: ${limit}`}>
            {limit === null ? n : `${n}/${limit}`}
          </span>
          {kindTag && <span className={`kind kind-${column.kind}`}>{kindTag}</span>}
        </h3>
        <span className="header-actions" onPointerDown={(e) => e.stopPropagation()}>
          <button className="icon" title="Ajustes de columna" onClick={onSettings}>⚙</button>
          <button className="icon" title="Eliminar columna"
            onClick={() => confirm(`¿Eliminar "${column.title}" y sus tarjetas?`) && onDelete()}>✕</button>
        </span>
      </div>
      <div ref={setDropRef} className="cards">
        <SortableContext items={column.cards.map((c) => `card-${c.id}`)} strategy={verticalListSortingStrategy}>
          {column.cards.map((c) => (
            <Card key={c.id} card={c} onOpen={() => onOpenCard(c)} />
          ))}
        </SortableContext>
      </div>
      {full ? (
        <div className="muted small wip-note">Límite de WIP alcanzado</div>
      ) : (
        <AddForm placeholder="Añadir tarjeta" withDetails onAdd={onAddCard} />
      )}
    </div>
  )
}
