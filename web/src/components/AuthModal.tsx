import { useState, useEffect, type FormEvent } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useLocale } from '../contexts/LocaleContext'

export default function AuthModal() {
  const { authModal, setAuthModal, login, register, loading } = useAuth()
  const { t } = useLocale()
  const [tab, setTab] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [pwd, setPwd] = useState('')
  const [err, setErr] = useState('')

  useEffect(() => {
    if (authModal) {
      setTab(authModal)
      setErr('')
      setEmail('')
      setName('')
      setPwd('')
    }
  }, [authModal])

  if (!authModal) return null

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setErr('')
    if (tab === 'login') {
      if (!email || !pwd) { setErr(t.auth.fillAll); return }
      const ok = await login(email, pwd)
      if (!ok) setErr(t.auth.loginFailed)
    } else {
      if (!email || !name || !pwd) { setErr(t.auth.fillAll); return }
      const ok = await register(email, name, pwd)
      if (!ok) setErr(t.auth.registerFailed)
    }
  }

  const inputCls = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-4 py-3 text-white placeholder-white/25 text-sm focus:outline-none focus:border-primary/40 transition-colors'

  return (
    <div
      className="fixed inset-0 z-100 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={() => setAuthModal(null)}
    >
      <div
        className="relative w-full max-w-md mx-4 bg-base-light border border-white/10 rounded-3xl p-8 shadow-[0_0_80px_rgba(0,212,255,0.12)]"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          className="absolute top-5 right-5 w-8 h-8 flex items-center justify-center rounded-full bg-white/5 text-white/40 hover:text-white/80 hover:bg-white/10 transition-colors text-lg leading-none"
          onClick={() => setAuthModal(null)}
        >
          x
        </button>

        <div className="flex items-center gap-2 mb-8">
          <span className="w-2 h-2 rounded-full bg-primary shadow-[0_0_10px_#00d4ff]" />
          <span className="text-white font-black text-xl tracking-tight">AgentHub</span>
        </div>

        <div className="flex gap-1 p-1 bg-white/5 rounded-2xl mb-7">
          {([['login', t.auth.loginTitle], ['register', t.auth.registerTitle]] as const).map(([k, label]) => (
            <button
              key={k}
              onClick={() => setTab(k as 'login' | 'register')}
              className={`flex-1 py-2.5 rounded-xl text-sm font-semibold transition-all ${
                tab === k
                  ? 'bg-primary text-base shadow-[0_0_15px_rgba(0,212,255,0.3)]'
                  : 'text-white/40 hover:text-white/70'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="email"
            placeholder={t.auth.email}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={inputCls}
          />
          {tab === 'register' && (
            <input
              type="text"
              placeholder={t.auth.name}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={inputCls}
            />
          )}
          <input
            type="password"
            placeholder={t.auth.password}
            value={pwd}
            onChange={(e) => setPwd(e.target.value)}
            className={inputCls}
          />
          {err && <p className="text-danger text-sm">{err}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-3.5 bg-primary text-base font-black rounded-xl hover:bg-[#00bfe8] transition-colors shadow-[0_0_20px_rgba(0,212,255,0.25)] mt-1 disabled:opacity-50"
          >
            {loading ? t.auth.loading : tab === 'login' ? t.auth.loginBtn : t.auth.registerBtn}
          </button>
        </form>

        <p className="mt-5 text-center text-white/30 text-sm">
          {tab === 'login' ? (
            <>
              {t.auth.noAccount}{' '}
              <button onClick={() => setTab('register')} className="text-primary hover:underline ml-1">
                {t.auth.goRegister}
              </button>
            </>
          ) : (
            <>
              {t.auth.hasAccount}{' '}
              <button onClick={() => setTab('login')} className="text-primary hover:underline ml-1">
                {t.auth.goLogin}
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  )
}
