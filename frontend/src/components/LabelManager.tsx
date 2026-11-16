import { useState } from 'react'
import { api } from '../api'
import { LABEL_COLORS, type Label } from '../types'
import { useToast } from './Toast'

interface Props {
  boardId: number
  labels: Label[]
  onClose: () => void
  onChange: () => void
}

function Palette({ value, onPick }: { value: string; onPick: (color: string) => void }) {
  return (
    <div className="palette" role="radiogroup" aria-label="Color">
      {LABEL_COLORS.map((c) => (
        <button key={c.value} type="button" role="radio" aria-checked={value.toLowerCase() === c.value}
          className={`swatch ${value.toLowerCase() === c.value ? 'selected' : ''}`}
          style={{ background: c.value }} title={c.name} aria-label={c.name} onClick={() => onPick(c.value)} />
      ))}
    </div>
  )
}

export default function LabelManager({ boardId, labels, onClose, onChange }: Props) {
  const toast = useToast()
  const [name, setName] = useState('')
  const [color, setColor] = useState(LABEL_COLORS[5].value)
  const [recoloring, setRecoloring] = useState<number | null>(null)

  const run = async (fn: () => Promise<unknown>) => {
    try { await fn() } catch (e) { toast((e as Error).message) }
    onChange()
  }

  return (
    <div className="overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>Etiquetas</h3>
        <ul className="label-list">
          {labels.map((l) => (
            <li key={l.id} className="label-item">
              <div className="label-row">
                <button type="button" className="swatch selected" style={{ background: l.color }} title="Cambiar color"
                  onClick={() => setRecoloring(recoloring === l.id ? null : l.id)} />
                <span className="chip" style={{ background: l.color }}>{l.name}</span>
                <span className="spacer" />
                <button className="icon" title="Eliminar"
                  onClick={() => confirm(`¿Eliminar la etiqueta "${l.name}"?`) && run(() => api.deleteLabel(l.id))}>✕</button>
              </div>
              {recoloring === l.id && (
                <Palette value={l.color} onPick={(c) => { setRecoloring(null); run(() => api.updateLabel(l.id, l.name, c)) }} />
              )}
            </li>
          ))}
          {labels.length === 0 && <li className="muted">Sin etiquetas todavía</li>}
        </ul>

        <form className="new-label" onSubmit={(e) => { e.preventDefault(); if (name.trim()) { run(() => api.createLabel(boardId, name.trim(), color)); setName('') } }}>
          <div className="muted small">Nueva etiqueta</div>
          <Palette value={color} onPick={setColor} />
          <div className="row">
            <input placeholder="Nombre" value={name} onChange={(e) => setName(e.target.value)} />
            <button className="primary" disabled={!name.trim()}>Añadir</button>
          </div>
        </form>
        <div className="row"><button onClick={onClose}>Cerrar</button></div>
      </div>
    </div>
  )
}
