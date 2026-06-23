import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { getAgent, listAgents } from '../api/agent'
import { useAuth } from '../contexts/AuthContext'
import type { Agent } from '../types'

export default function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, setAuthModal } = useAuth()
  const [agent, setAgent] = useState<Agent | null>(null)
  const [related, setRelated] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    getAgent(Number(id))
      .then((a) => {
        setAgent(a)
        if (a.category) {
          listAgents({ category: a.category, limit: 3 })
            .then((r) => setRelated(r.agents.filter((x) => x.id !== a.id).slice(0, 2)))
            .catch(() => {})
        }
      })
      .catch(() => setAgent(null))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-white/30">Loading...</p>
      </div>
    )
  }

  if (!agent) {
    return (
      <div className="min-h-screen flex items-center justify-center text-center px-6">
        <div>
          <div className="text-6xl mb-4">🔍</div>
          <p className="text-white/40 mb-6">Agent not found</p>
          <Link to="/agents" className="text-primary hover:underline">&larr; Back to Market</Link>
        </div>
      </div>
    )
  }

  const color = agent.color || '#00d4ff'
  const tags = agent.tags || []

  const handleChat = () => {
    if (!user) {
      setAuthModal('login')
      return
    }
    navigate(`/chat/${agent.id}`)
  }

  return (
    <div className="min-h-screen py-12 px-6">
      <div className="max-w-6xl mx-auto">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-1.5 text-white/35 hover:text-white/75 mb-9 transition-colors text-sm"
        >
          &larr; Back
        </button>

        <div className="grid lg:grid-cols-3 gap-8">
          {/* Left */}
          <div className="lg:col-span-2 space-y-5">
            <div className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-8">
              <div className="flex items-start gap-5 mb-6">
                <div
                  className="w-16 h-16 rounded-2xl flex items-center justify-center text-3xl shrink-0"
                  style={{ background: `${color}18`, border: `1px solid ${color}30` }}
                >
                  {agent.icon || '🤖'}
                </div>
                <div>
                  <div className="flex flex-wrap items-center gap-2 mb-1">
                    <h1 className="text-2xl font-black text-white">{agent.name}</h1>
                    {agent.category && (
                      <span
                        className="px-2 py-0.5 rounded-full text-xs"
                        style={{ background: `${color}18`, color, border: `1px solid ${color}28` }}
                      >
                        {agent.category}
                      </span>
                    )}
                  </div>
                  <p className="text-white/55">{agent.description}</p>
                  <div className="flex flex-wrap items-center gap-4 mt-2">
                    <span className="text-white/30 text-sm">{(agent.call_count / 1000).toFixed(0)}k calls</span>
                    <span className="text-white/30 text-sm">{agent.model_name || 'LLM'}</span>
                  </div>
                </div>
              </div>

              {tags.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {tags.map((t) => (
                    <span key={t} className="px-3 py-1 bg-white/[0.04] border border-white/[0.07] rounded-full text-white/40 text-sm">
                      {t}
                    </span>
                  ))}
                </div>
              )}
            </div>

            <div className="grid grid-cols-3 gap-4">
              {[
                { label: 'Model', val: agent.model_name || '-' },
                { label: 'Temperature', val: String(agent.temperature) },
                { label: 'Calls', val: `${(agent.call_count / 1000).toFixed(1)}k` },
              ].map((s) => (
                <div key={s.label} className="bg-white/[0.03] border border-white/[0.07] rounded-xl p-4 text-center">
                  <div className="text-xl font-black text-white font-mono">{s.val}</div>
                  <div className="text-white/35 text-xs mt-1">{s.label}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Right */}
          <div className="space-y-5">
            <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 sticky top-24">
              <div className="text-center mb-6">
                <div className="flex items-center justify-center gap-2 mb-1">
                  <span className="text-accent text-xl">&#9670;</span>
                  <span className="text-accent font-black text-4xl font-mono">{agent.cost}</span>
                </div>
                <p className="text-white/35 text-sm">credits / call</p>
              </div>

              <button
                onClick={handleChat}
                className="w-full py-4 rounded-xl font-black text-base transition-all bg-primary text-base hover:bg-[#00bfe8] shadow-[0_0_25px_rgba(0,212,255,0.25)]"
              >
                {user ? '▶ Start Chat' : 'Login to Chat'}
              </button>

              {!user && (
                <p className="mt-4 text-center text-white/25 text-xs">
                  <button onClick={() => setAuthModal('login')} className="text-primary hover:underline">Login</button>
                  {' or '}
                  <button onClick={() => setAuthModal('register')} className="text-primary hover:underline">Register</button>
                  {' to start'}
                </p>
              )}

              <div className="mt-6 pt-5 border-t border-white/[0.06] space-y-2.5">
                {[
                  { label: 'Status', val: agent.status },
                  { label: 'Model', val: agent.model_name || '-' },
                  { label: 'Max Tokens', val: String(agent.max_tokens) },
                ].map((r) => (
                  <div key={r.label} className="flex justify-between text-sm">
                    <span className="text-white/35">{r.label}</span>
                    <span className="text-white/65 font-mono">{r.val}</span>
                  </div>
                ))}
              </div>
            </div>

            {related.length > 0 && (
              <div className="bg-white/[0.02] border border-white/[0.07] rounded-2xl p-5">
                <p className="text-white/30 text-xs font-mono mb-3 tracking-wider">// Related</p>
                <div className="space-y-2">
                  {related.map((r) => (
                    <Link
                      key={r.id}
                      to={`/agents/${r.id}`}
                      className="flex items-center gap-3 px-2 py-2.5 rounded-xl hover:bg-white/[0.05] transition-colors"
                    >
                      <span className="text-xl">{r.icon || '🤖'}</span>
                      <div className="min-w-0">
                        <p className="text-white/75 text-sm font-medium truncate">{r.name}</p>
                        <p className="text-white/30 text-xs">{r.cost} credits</p>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
