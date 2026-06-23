import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import Toast from '../components/Toast'

const nav = [
  { to: '/admin', label: 'Overview', icon: '⊡' },
  { to: '/admin/agents', label: 'Agents', icon: '⬡' },
  { to: '/admin/knowledge-bases', label: 'Knowledge Bases', icon: '◈' },
]

export default function AdminLayout() {
  const { user, logout } = useAuth()
  const location = useLocation()

  const isActive = (to: string) => {
    if (to === '/admin') return location.pathname === '/admin'
    return location.pathname.startsWith(to)
  }

  return (
    <div className="min-h-screen bg-base text-white flex">
      <div className="fixed inset-0 pointer-events-none grid-bg-admin opacity-50" />

      <div className="w-56 shrink-0 bg-[#070a18] border-r border-white/[0.06] flex flex-col h-screen fixed left-0 top-0 z-20">
        <div className="h-16 flex items-center px-5 border-b border-white/[0.05] gap-2">
          <span className="w-2 h-2 rounded-full bg-secondary shadow-[0_0_8px_#7c3aed]" />
          <span className="font-black text-white tracking-tight">AgentHub</span>
          <span className="ml-1 px-1.5 py-0.5 bg-secondary/20 text-secondary text-[10px] font-black rounded">
            ADMIN
          </span>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1">
          {nav.map(({ to, label, icon }) => (
            <Link
              key={to}
              to={to}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all text-left ${
                isActive(to)
                  ? 'bg-secondary/[0.22] text-secondary font-semibold border border-secondary/20'
                  : 'text-white/40 hover:text-white/75 hover:bg-white/[0.04]'
              }`}
            >
              <span>{icon}</span>
              {label}
            </Link>
          ))}
          <Link
            to="/"
            className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-white/40 hover:text-white/75 hover:bg-white/[0.04] transition-all text-left"
          >
            <span>&larr;</span>
            Back to Site
          </Link>
        </nav>

        <div className="p-4 border-t border-white/[0.05]">
          <div className="flex items-center gap-2.5 px-1">
            <div className="w-7 h-7 rounded-full bg-secondary flex items-center justify-center text-white text-xs font-black">
              {user?.name?.slice(0, 1).toUpperCase() || 'A'}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white/65 text-sm truncate">{user?.name || 'Admin'}</p>
              <p className="text-white/25 text-xs truncate">{user?.email}</p>
            </div>
            <button onClick={logout} className="text-white/25 hover:text-white/60 text-xs transition-colors">
              Logout
            </button>
          </div>
        </div>
      </div>

      <div className="flex-1 ml-56 relative min-h-screen overflow-y-auto">
        <Outlet />
      </div>

      <Toast />
    </div>
  )
}
