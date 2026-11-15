import { useCallback, useEffect, useState } from 'react'
import { api, ApiError } from '../api'
import type { BoardDetail, Column, User } from '../types'
import { BoardCtx } from '../context'
import { useToast } from '../components/Toast'
import Board from '../components/Board'
import LabelManager from '../components/LabelManager'
import BoardTabs from '../components/BoardTabs'

export default function BoardPage({ id }: { id: number }) {
  const toast = useToast()
  const [board, setBoard] = useState<BoardDetail | null>(null)
  const [users, setUsers] = useState<User[]>([])
  const [showLabels, setShowLabels] = useState(false)

  const reload = useCallback(() => {
    api.board(id).then(setBoard).catch((e: ApiError) => {
      toast(e.message)
      if (e.status === 404) location.hash = '#/'
    })
  }, [id, toast])

  useEffect(() => {
    reload()
    api.users().then(setUsers).catch(() => {})
  }, [reload])

  const setColumns: React.Dispatch<React.SetStateAction<Column[]>> = (u) =>
    setBoard((b) => (b ? { ...b, columns: typeof u === 'function' ? u(b.columns) : u } : b))

  if (!board) return <main className="page"><p className="muted">Cargando…</p></main>

  return (
    <BoardCtx.Provider value={{ labels: board.labels, users }}>
      <BoardTabs boardId={id} name={board.name} active="board"
        actions={<button onClick={() => setShowLabels(true)}>Etiquetas</button>} />
      <Board boardId={id} columns={board.columns} setColumns={setColumns} reload={reload} />
      {showLabels && <LabelManager boardId={id} labels={board.labels} onClose={() => setShowLabels(false)} onChange={reload} />}
    </BoardCtx.Provider>
  )
}
