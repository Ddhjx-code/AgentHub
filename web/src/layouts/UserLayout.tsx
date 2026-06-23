import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useLocale } from '../contexts/LocaleContext'
import AuthModal from '../components/AuthModal'
import Toast from '../components/Toast'

export default function UserLayout() {
  const { user, logout, setAuthModal } = useAuth()
  const { t, locale, toggleLocale } = useLocale()
  const location = useLocation()

  const links = [
    { to: '/', label: t.nav.home },
    { to: '/agents', label: t.nav.agents },
  ]

  return (
    <div className="min-h-screen bg-base text-white">
      <div className="fixed inset-0 pointer-events-none grid-bg opacity-60" />
      <div className="fixed top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-primary/[0.03] blur-[120px] pointer-events-none rounded-full" />

      <nav className="fixed top-0 left-0 right-0 z-50 border-b border-white/[0.06] bg-base/75 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between gap-6">
          <Link to="/" className="flex items-center gap-2 shrink-0">
            <span className="w-2 h-2 rounded-full bg-primary shadow-[0_0_10px_#00d4ff]" />
            <span className="font-black text-lg tracking-tight">AgentHub</span>
          </Link>

          <div className="hidden md:flex items-center gap-8">
            {links.map(({ to, label }) => (
              <Link
                key={to}
                to={to}
                className={`text-sm font-medium transition-colors ${
                  location.pathname === to ? 'text-primary' : 'text-white/45 hover:text-white/85'
                }`}
              >
                {label}
              </Link>
            ))}
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <button
              onClick={toggleLocale}
              className="px-2.5 py-1 bg-white/[0.05] border border-white/[0.08] rounded-full text-xs font-mono text-white/50 hover:text-white/80 hover:bg-white/[0.08] transition-colors"
            >
              {locale === 'zh' ? 'EN' : '中'}
            </button>

            {user ? (
              <>
                <Link
                  to="/dashboard"
                  className="flex items-center gap-2 px-3 py-1.5 bg-white/[0.05] hover:bg-white/10 border border-white/[0.07] rounded-full transition-colors"
                >
                  <div className="w-6 h-6 rounded-full bg-primary flex items-center justify-center text-base text-xs font-black">
                    {user.name.slice(0, 2).toUpperCase()}
                  </div>
                  <span className="hidden sm:block text-sm text-white/75">{user.name}</span>
                </Link>
                {user.role === 'admin' && (
                  <Link
                    to="/admin"
                    className="px-3 py-1.5 bg-secondary/15 border border-secondary/25 text-secondary text-xs font-bold rounded-full hover:bg-secondary/25 transition-colors"
                  >
                    {t.nav.admin}
                  </Link>
                )}
                <button
                  onClick={logout}
                  className="hidden md:block text-white/35 hover:text-white/65 text-sm transition-colors"
                >
                  {t.nav.logout}
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={() => setAuthModal('login')}
                  className="text-white/50 hover:text-white text-sm transition-colors"
                >
                  {t.nav.login}
                </button>
                <button
                  onClick={() => setAuthModal('register')}
                  className="px-4 py-2 bg-primary text-base text-sm font-black rounded-full hover:bg-[#00bfe8] transition-colors shadow-[0_0_15px_rgba(0,212,255,0.2)]"
                >
                  {t.nav.register}
                </button>
              </>
            )}
          </div>
        </div>
      </nav>

      <main className="relative pt-16">
        <Outlet />
      </main>

      <Toast />
      <AuthModal />
    </div>
  )
}
