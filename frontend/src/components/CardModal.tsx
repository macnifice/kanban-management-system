import { useContext, useEffect, useState } from 'react'
import { api } from '../api'
import { BoardCtx } from '../context'
import { PRIORITIES, type Card, type HistoryEvent, type Priority } from '../types'

interface Props {
  card: Card
  onSave: (patch: { title: string; description: string; priority: Priority; assignee_id: number | null; label_ids: number[] }) => void
  onDelete: () => void
  onClose: () => void
}

export default function CardModal({ card, onSave, onDelete, onClose }: Props) {
  const { labels, users } = useContext(BoardCtx)
  const [title, setTitle] = useState(card.title)
  const [description, setDescription] = useState(card.description)
  const [priority, setPriority] = useState<Priority>(card.priority)
  const [assignee, setAssignee] = useState<number | null>(card.assignee_id)
  const [labelIds, setLabelIds] = useState<number[]>(card.label_ids)
  const [history, setHistory] = useState<HistoryEvent[] | null>(null)

  useEffect(() => {
    api.history(card.id).then(setHistory).catch(() => setHistory([]))
  }, [card.id])

  const toggleLabel = (id: number) =>
    setLabelIds((xs) => (xs.includes(id) ? xs.filter((x) => x !== id) : [...xs, id]))

  return (
    <div className="overlay" onClick={onClose}>
      <div className="modal wide" onClick={(e) => e.stopPropagation()}>
        <input value={title} onChange={(e) => setTitle(e.target.value)} />
        <textarea rows={5} value={description} placeholder="Descripción…" onChange={(e) => setDescription(e.target.value)} />

        <div className="grid2">
          <label>Prioridad
            <select value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
              {PRIORITIES.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </label>
          <label>Asignado a
            <select value={assignee ?? ''} onChange={(e) => setAssignee(e.target.value ? Number(e.target.value) : null)}>
              <option value="">Sin asignar</option>
              {users.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
            </select>
          </label>
        </div>

        {labels.length > 0 && (
          <div>
            <div className="muted small">Etiquetas</div>
            <div className="chips">
              {labels.map((l) => (
                <button key={l.id} type="button" className={`chip chip-btn ${labelIds.includes(l.id) ? 'on' : ''}`}
                  style={{ background: labelIds.includes(l.id) ? l.color : 'transparent', color: labelIds.includes(l.id) ? '#fff' : l.color, borderColor: l.color }}
                  onClick={() => toggleLabel(l.id)}>{l.name}</button>
              ))}
            </div>
          </div>
        )}

        <div className="muted small">Creada el {new Date(card.created_at).toLocaleString()}</div>

        <div className="history">
          <div className="muted small">Historial de movimientos</div>
          {history === null && <div className="muted small">Cargando…</div>}
          <ul>
            {history?.map((e) => (
              <li key={e.id}>
                <span className="muted">{new Date(e.moved_at).toLocaleString()}</span>{' '}
                {e.from_column ? <>de <b>{e.from_column}</b> a <b>{e.to_column}</b></> : <>creada en <b>{e.to_column}</b></>}
                {e.user && <span className="muted"> · {e.user}</span>}
              </li>
            ))}
          </ul>
        </div>

        <div className="row">
          <button className="primary" disabled={!title.trim()}
            onClick={() => { onSave({ title: title.trim(), description, priority, assignee_id: assignee, label_ids: labelIds }); onClose() }}>
            Guardar
          </button>
          <button onClick={onClose}>Cancelar</button>
          <button className="danger" onClick={() => { if (confirm('¿Eliminar esta tarjeta?')) { onDelete(); onClose() } }}>Eliminar</button>
        </div>
      </div>
    </div>
  )
}
