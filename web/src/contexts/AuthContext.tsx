import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react'
import type { User } from '../types'
import * as authApi from '../api/auth'

interface AuthState {
  user: User | null
  token: string | null
  loading: boolean
  authModal: 'login' | 'register' | null
  toast: { msg: string; type: 'success' | 'error' | 'info' } | null
  login: (email: string, password: string) => Promise<boolean>
  register: (email: string, name: string, password: string) => Promise<boolean>
  logout: () => void
  setAuthModal: (modal: 'login' | 'register' | null) => void
  flash: (msg: string, type?: 'success' | 'error' | 'info') => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => {
    const stored = localStorage.getItem('user')
    return stored ? JSON.parse(stored) : null
  })
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [loading, setLoading] = useState(false)
  const [authModal, setAuthModal] = useState<'login' | 'register' | null>(null)
  const [toast, setToast] = useState<{ msg: string; type: 'success' | 'error' | 'info' } | null>(null)

  const flash = useCallback((msg: string, type: 'success' | 'error' | 'info' = 'success') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3200)
  }, [])

  const login = useCallback(async (email: string, password: string): Promise<boolean> => {
    setLoading(true)
    try {
      const data = await authApi.login(email, password)
      setToken(data.token)
      setUser(data.user)
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      setAuthModal(null)
      flash(`Welcome back, ${data.user.name}!`)
      return true
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Login failed', 'error')
      return false
    } finally {
      setLoading(false)
    }
  }, [flash])

  const register = useCallback(async (email: string, name: string, password: string): Promise<boolean> => {
    setLoading(true)
    try {
      await authApi.register(email, name, password)
      const data = await authApi.login(email, password)
      setToken(data.token)
      setUser(data.user)
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      setAuthModal(null)
      flash('Registration successful!')
      return true
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Registration failed', 'error')
      return false
    } finally {
      setLoading(false)
    }
  }, [flash])

  const logout = useCallback(() => {
    setUser(null)
    setToken(null)
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    flash('Logged out', 'info')
  }, [flash])

  useEffect(() => {
    if (token && !user) {
      authApi.getProfile()
        .then((u) => {
          setUser(u)
          localStorage.setItem('user', JSON.stringify(u))
        })
        .catch(() => {
          setToken(null)
          localStorage.removeItem('token')
          localStorage.removeItem('user')
        })
    }
  }, [token, user])

  return (
    <AuthContext.Provider
      value={{ user, token, loading, authModal, toast, login, register, logout, setAuthModal, flash }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
