import { useState } from 'react'
import { KINDS, type Column, type Kind } from '../types'

interface Props {
  column: Column
  onSave: (patch: { title: string; wip_limit: number | null; kind: Kind }) => void
  onClose: () => void
}

export default function ColumnSettings({ column, onSave, onClose }: Props) {
  const [title, setTitle] = useState(column.title)
  const [wip, setWip] = useState(column.wip_limit === null ? '' : String(column.wip_limit))
  const [kind, setKind] = useState<Kind>(column.kind)
  const wipNum = wip.trim() === '' ? null : Number(wip)
  const valid = title.trim() !== '' && (wipNum === null || (Number.isInteger(wipNum) && wipNum >= 1))

  return (
    <div className="overlay" onClick={onClose}>
      <form className="modal" onClick={(e) => e.stopPropagation()}
        onSubmit={(e) => { e.preventDefault(); if (valid) { onSave({ title: title.trim(), wip_limit: wipNum, kind }); onClose() } }}>
        <h3>Ajustes de columna</h3>
        <label>Nombre<input autoFocus value={title} onChange={(e) => setTitle(e.target.value)} /></label>
        <label>Límite de WIP
          <input type="number" min={1} placeholder="Sin límite" value={wip} onChange={(e) => setWip(e.target.value)} />
        </label>
        {column.cards.length > 0 && wipNum !== null && wipNum < column.cards.length && (
          <div className="form-error">La columna ya tiene {column.cards.length} tarjetas: quedará por encima del límite y no admitirá más.</div>
        )}
        <label>Tipo de columna
          <select value={kind} onChange={(e) => setKind(e.target.value as Kind)}>
            {KINDS.map((k) => <option key={k.value} value={k.value}>{k.label}</option>)}
          </select>
        </label>
        <div className="muted small">{KINDS.find((k) => k.value === kind)?.hint}</div>
        <div className="row">
          <button className="primary" disabled={!valid}>Guardar</button>
          <button type="button" onClick={onClose}>Cancelar</button>
        </div>
      </form>
    </div>
  )
}
