import { useCallback, useEffect, useState } from 'react'
import { api, getToken, setToken, setUnauthorizedHandler } from './api'
import type { Board, User } from './types'
import LoginPage from './pages/LoginPage'
import BoardPage from './pages/BoardPage'
import MetricsPage from './pages/MetricsPage'
import Sidebar from './components/Sidebar'

type Route = { page: 'boards' } | { page: 'board' | 'metrics'; id: number }

function parse(hash: string): Route {
  const parts = hash.replace(/^#/, '').split('/').filter(Boolean)
  if (parts[0] === 'boards' && /^\d+$/.test(parts[1] ?? '')) {
    return { page: parts[2] === 'metrics' ? 'metrics' : 'board', id: Number(parts[1]) }
  }
  return { page: 'boards' }
}

export default function App() {
  const [user, setUser] = useState<User | null>(null)
  const [ready, setReady] = useState(false)
  const [route, setRoute] = useState<Route>(() => parse(location.hash))
  const [boards, setBoards] = useState<Board[] | null>(null)
  const [sidebarOpen, setSidebarOpen] = useState(true)

  const loadBoards = useCallback(
    () => api.boards().then(setBoards).catch(() => setBoards([])),
    [],
  )

  useEffect(() => {
    const onHash = () => setRoute(parse(location.hash))
    window.addEventListener('hashchange', onHash)
    setUnauthorizedHandler(() => { setToken(null); setUser(null) })
    if (getToken()) {
      api.me().then(setUser).catch(() => setToken(null)).finally(() => setReady(true))
    } else {
      setReady(true)
    }
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  // Carga la lista al iniciar sesión y la refresca en cada cambio de ruta (contadores del sidebar).
  useEffect(() => {
    if (user) loadBoards()
  }, [route, user, loadBoards])

  // "#/" ya no tiene vista propia: abre el primer tablero disponible.
  useEffect(() => {
    if (route.page === 'boards' && boards && boards.length > 0) location.replace(`#/boards/${boards[0].id}`)
  }, [route, boards])

  if (!ready) return null
  if (!user) return <LoginPage onLogin={setUser} />

  const logout = () => { setToken(null); setUser(null) }

  return (
    <div className="app">
      <header>
        <button className="menu-btn" aria-label="Mostrar u ocultar tableros" onClick={() => setSidebarOpen((o) => !o)}>☰</button>
        <a href="#/" className="brand">Kanban</a>
        <span className="spacer" />
        <span className="who">{user.name}</span>
        <button onClick={logout}>Salir</button>
      </header>
      <div className="layout">
        {sidebarOpen && (
          <Sidebar boards={boards} onChanged={loadBoards}
            currentId={route.page === 'boards' ? null : route.id}
            tab={route.page === 'metrics' ? 'metrics' : 'board'} />
        )}
        <div className="content">
          {route.page === 'boards' && (
            <main className="page empty">
              {boards === null ? <p className="muted">Cargando…</p> : (
                <>
                  <h2>Bienvenido</h2>
                  <p className="muted">Aún no tienes tableros. Crea el primero con el botón ＋ del panel izquierdo.</p>
                </>
              )}
            </main>
          )}
          {route.page === 'board' && <BoardPage key={route.id} id={route.id} />}
          {route.page === 'metrics' && <MetricsPage key={route.id} id={route.id} />}
        </div>
      </div>
    </div>
  )
}
