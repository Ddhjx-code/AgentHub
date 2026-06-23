import { useAuth } from '../contexts/AuthContext'

const palette: Record<string, string> = {
  success: 'border-primary bg-primary/10 text-primary',
  error: 'border-danger bg-danger/10 text-danger',
  info: 'border-white/20 bg-white/5 text-white/60',
}

export default function Toast() {
  const { toast } = useAuth()
  if (!toast) return null

  return (
    <div
      className={`fixed bottom-6 right-6 z-200 flex items-center gap-3 px-5 py-3 rounded-2xl border backdrop-blur-xl animate-fadeInUp ${palette[toast.type] || palette.success}`}
    >
      <span className="text-sm font-semibold">{toast.msg}</span>
    </div>
  )
}
