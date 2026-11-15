import { useState } from 'react'
import { PRIORITIES, type Priority } from '../types'

interface Props {
  placeholder: string
  withDetails?: boolean
  onAdd: (title: string, description: string, priority: Priority) => void
}

export default function AddForm({ placeholder, withDetails, onAdd }: Props) {
  const [open, setOpen] = useState(false)
  const [value, setValue] = useState('')
  const [desc, setDesc] = useState('')
  const [priority, setPriority] = useState<Priority>('medium')

  if (!open) {
    return (
      <button className="add-btn" onClick={() => setOpen(true)}>
        + {placeholder}
      </button>
    )
  }
  const close = () => { setValue(''); setDesc(''); setPriority('medium'); setOpen(false) }
  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (value.trim()) onAdd(value.trim(), desc, priority)
    close()
  }
  return (
    <form className="add-form" onSubmit={submit}>
      <input autoFocus value={value} onChange={(e) => setValue(e.target.value)} placeholder="Título"
        onKeyDown={(e) => e.key === 'Escape' && close()} />
      {withDetails && (
        <>
          <textarea rows={3} value={desc} onChange={(e) => setDesc(e.target.value)} placeholder="Descripción (opcional)"
            onKeyDown={(e) => e.key === 'Escape' && close()} />
          <select value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
            {PRIORITIES.map((p) => <option key={p.value} value={p.value}>Prioridad: {p.label}</option>)}
          </select>
        </>
      )}
      <div className="row">
        <button type="submit" className="primary">Añadir</button>
        <button type="button" onClick={close}>Cancelar</button>
      </div>
    </form>
  )
}
