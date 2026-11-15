import { useState } from 'react'
import { api } from '../api'
import type { Board } from '../types'
import { useToast } from './Toast'

interface Props {
  boards: Board[] | null
  currentId: number | null
  tab: 'board' | 'metrics' // al cambiar de tablero se conserva la pestaña activa
  onChanged: () => Promise<void>
}

export default function Sidebar({ boards, currentId, tab, onChanged }: Props) {
  const toast = useToast()
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [editing, setEditing] = useState<{ id: number; name: string; description: string } | null>(null)

  const close = () => { setCreating(false); setName('') }

  const create = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || busy) return
    setBusy(true)
    try {
      const b = await api.createBoard(name.trim(), '')
      close()
      await onChanged()
      location.hash = `#/boards/${b.id}`
    } catch (err) {
      toast((err as Error).message)
    } finally {
      setBusy(false)
    }
  }

  const saveEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editing || !editing.name.trim()) return
    try {
      await api.updateBoard(editing.id, editing.name.trim(), editing.description)
      setEditing(null)
    } catch (err) {
      toast((err as Error).message)
    }
    await onChanged()
  }

  const remove = async (b: Board) => {
    if (!confirm(`¿Eliminar el tablero "${b.name}" con sus columnas y tarjetas?`)) return
    try {
      await api.deleteBoard(b.id)
    } catch (err) {
      toast((err as Error).message)
    }
    await onChanged()
    if (b.id === currentId) location.hash = '#/' // App redirige al primer tablero que quede
  }

  return (
    <aside className="sidebar" aria-label="Tableros">
      <div className="sidebar-head">
        <span className="sidebar-title">Tableros</span>
        <button className="icon" title="Nuevo tablero" aria-label="Nuevo tablero" onClick={() => setCreating(true)}>＋</button>
      </div>

      {creating && (
        <form className="sidebar-new" onSubmit={create}>
          <input autoFocus placeholder="Nombre del tablero" value={name} onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === 'Escape' && close()} />
          <div className="row">
            <button className="primary" disabled={!name.trim() || busy}>Crear</button>
            <button type="button" onClick={close}>Cancelar</button>
          </div>
        </form>
      )}

      <nav className="sidebar-list">
        {boards === null && <span className="muted small">Cargando…</span>}
        {boards?.length === 0 && !creating && <span className="muted small">Sin tableros. Crea el primero con ＋</span>}
        {boards?.map((b) => (
          <div key={b.id} className={`sidebar-item ${b.id === currentId ? 'active' : ''}`}>
            <a href={`#/boards/${b.id}${tab === 'metrics' ? '/metrics' : ''}`} className="sidebar-link"
              aria-current={b.id === currentId ? 'page' : undefined}>
              <span className="sidebar-name">{b.name}</span>
              <span className="sidebar-count">{b.card_count}</span>
            </a>
            <span className="sidebar-actions">
              <button className="icon" title="Editar tablero" aria-label={`Editar ${b.name}`}
                onClick={() => setEditing({ id: b.id, name: b.name, description: b.description })}>✎</button>
              <button className="icon" title="Eliminar tablero" aria-label={`Eliminar ${b.name}`} onClick={() => remove(b)}>✕</button>
            </span>
          </div>
        ))}
      </nav>

      {editing && (
        <div className="overlay" onClick={() => setEditing(null)}>
          <form className="modal" onClick={(e) => e.stopPropagation()} onSubmit={saveEdit}>
            <h3>Editar tablero</h3>
            <input autoFocus placeholder="Nombre" value={editing.name} onChange={(e) => setEditing({ ...editing, name: e.target.value })} />
            <textarea rows={3} placeholder="Descripción (opcional)" value={editing.description}
              onChange={(e) => setEditing({ ...editing, description: e.target.value })} />
            <div className="row">
              <button className="primary" disabled={!editing.name.trim()}>Guardar</button>
              <button type="button" onClick={() => setEditing(null)}>Cancelar</button>
            </div>
          </form>
        </div>
      )}
    </aside>
  )
}
