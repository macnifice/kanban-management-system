import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'

type Kind = 'error' | 'info'
interface Item { id: number; text: string; kind: Kind }
type Notify = (text: string, kind?: Kind) => void

const ToastCtx = createContext<Notify>(() => {})
export const useToast = () => useContext(ToastCtx)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<Item[]>([])
  const notify = useCallback<Notify>((text, kind = 'error') => {
    const id = Date.now() + Math.random()
    setItems((xs) => [...xs, { id, text, kind }])
    setTimeout(() => setItems((xs) => xs.filter((x) => x.id !== id)), 4500)
  }, [])
  return (
    <ToastCtx.Provider value={notify}>
      {children}
      <div className="toasts">
        {items.map((t) => (
          <div key={t.id} className={`toast ${t.kind}`} onClick={() => setItems((xs) => xs.filter((x) => x.id !== t.id))}>
            {t.text}
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  )
}
