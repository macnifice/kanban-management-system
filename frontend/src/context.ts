import { createContext } from 'react'
import type { Label, User } from './types'

export const BoardCtx = createContext<{ labels: Label[]; users: User[] }>({ labels: [], users: [] })
