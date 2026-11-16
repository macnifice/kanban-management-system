import { useState } from 'react'
import { api, setToken } from '../api'
import type { User } from '../types'

export default function LoginPage({ onLogin }: { onLogin: (u: User) => void }) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      const r = mode === 'login' ? await api.login(email, password) : await api.register(email, name, password)
      setToken(r.token)
      onLogin(r.user)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login-wrap">
      <form className="login" onSubmit={submit}>
        <h1>Kanban</h1>
        <p className="muted">{mode === 'login' ? 'Inicia sesión para continuar' : 'Crea tu cuenta'}</p>
        {mode === 'register' && <input placeholder="Nombre" value={name} onChange={(e) => setName(e.target.value)} required />}
        <input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input type="password" placeholder="Contraseña (mín. 6)" value={password} minLength={mode === 'register' ? 6 : undefined}
          onChange={(e) => setPassword(e.target.value)} required />
        {error && <div className="form-error">{error}</div>}
        <button className="primary" disabled={busy}>{mode === 'login' ? 'Entrar' : 'Registrarme'}</button>
        <button type="button" className="link" onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setError('') }}>
          {mode === 'login' ? '¿No tienes cuenta? Regístrate' : '¿Ya tienes cuenta? Inicia sesión'}
        </button>
      </form>
    </div>
  )
}
