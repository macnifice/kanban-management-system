import type { ReactNode } from 'react'

interface Props {
  boardId: number
  name: string
  active: 'board' | 'metrics'
  actions?: ReactNode // botones a la derecha (p. ej. Etiquetas)
}

// Cabecera común del tablero: volver, nombre y pestañas Tablero / Métricas.
export default function BoardTabs({ boardId, name, active, actions }: Props) {
  return (
    <div className="board-top">
      <div className="board-top-row">
        <h2>{name}</h2>
        <span className="spacer" />
        {actions}
      </div>
      <nav className="tabs" role="tablist">
        <a role="tab" aria-selected={active === 'board'} className={`tab ${active === 'board' ? 'active' : ''}`}
          href={`#/boards/${boardId}`}>Tablero</a>
        <a role="tab" aria-selected={active === 'metrics'} className={`tab ${active === 'metrics' ? 'active' : ''}`}
          href={`#/boards/${boardId}/metrics`}>Métricas</a>
      </nav>
    </div>
  )
}
