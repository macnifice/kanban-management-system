import { useContext } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { PRIORITIES, type Card as CardT } from '../types'
import { BoardCtx } from '../context'

export default function Card({ card, onOpen }: { card: CardT; onOpen: () => void }) {
  const { labels, users } = useContext(BoardCtx)
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `card-${card.id}`,
    data: { type: 'card', card },
  })
  const pr = PRIORITIES.find((p) => p.value === card.priority)!
  const assignee = users.find((u) => u.id === card.assignee_id)
  const cardLabels = labels.filter((l) => card.label_ids.includes(l.id))

  return (
    <div
      ref={setNodeRef}
      className="card"
      style={{ transform: CSS.Transform.toString(transform), transition, opacity: isDragging ? 0.4 : 1, borderLeftColor: pr.color }}
      onClick={onOpen}
      {...attributes}
      {...listeners}
    >
      {cardLabels.length > 0 && (
        <div className="chips">
          {cardLabels.map((l) => <span key={l.id} className="chip" style={{ background: l.color }}>{l.name}</span>)}
        </div>
      )}
      <div className="card-title">{card.title}</div>
      {card.description && <div className="card-desc">{card.description}</div>}
      <div className="card-foot">
        <span className="prio" style={{ color: pr.color }}>● {pr.label}</span>
        <span className="muted">{new Date(card.created_at).toLocaleDateString()}</span>
        {assignee && <span className="avatar" title={assignee.name}>{assignee.name.charAt(0).toUpperCase()}</span>}
      </div>
    </div>
  )
}
