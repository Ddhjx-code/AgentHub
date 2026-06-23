import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { listAgents } from '../api/agent'
import { listConversations } from '../api/chat'
import type { Agent, Conversation } from '../types'

export default function Dashboard() {
  const { user, logout } = useAuth()
  const [agents, setAgents] = useState<Agent[]>([])
  const [conversations, setConversations] = useState<Conversation[]>([])

  useEffect(() => {
    listAgents({ limit: 4 }).then((r) => setAgents(r.agents)).catch(() => {})
    listConversations().then(setConversations).catch(() => {})
  }, [])

  if (!user) return null

  return (
    <div className="min-h-screen py-12 px-6">
      <div className="max-w-6xl mx-auto">
        <div className="flex items-start justify-between mb-10 gap-4">
          <div>
            <p className="text-primary text-xs font-mono tracking-widest uppercase mb-2">// Dashboard</p>
            <h1 className="text-3xl font-black text-white">Hello, {user.name}</h1>
            <p className="text-white/35 text-sm mt-1">{user.email}</p>
          </div>
          <button
            onClick={logout}
            className="shrink-0 px-4 py-2 bg-white/[0.04] border border-white/[0.08] rounded-xl text-white/40 text-sm hover:bg-white/[0.08] hover:text-white/65 transition-colors"
          >
            Logout
          </button>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          {[
            { label: 'Role', val: user.role, color: '#00d4ff' },
            { label: 'Status', val: user.status, color: '#10b981' },
            { label: 'Conversations', val: String(conversations.length), color: '#7c3aed' },
          ].map((s) => (
            <div key={s.label} className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-6">
              <p className="text-white/35 text-sm mb-3">{s.label}</p>
              <div className="text-3xl font-black font-mono" style={{ color: s.color }}>{s.val}</div>
            </div>
          ))}
        </div>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Recent conversations */}
          <div className="lg:col-span-2 bg-white/[0.03] border border-white/[0.07] rounded-2xl p-6">
            <p className="text-white/35 text-xs font-mono tracking-wider mb-5">// Recent Conversations</p>
            {conversations.length === 0 ? (
              <p className="text-white/25 text-sm py-4">No conversations yet</p>
            ) : (
              <div className="space-y-0.5">
                {conversations.slice(0, 10).map((conv) => (
                  <Link
                    key={conv.id}
                    to={`/chat/${conv.agent_id}?conversation=${conv.id}`}
                    className="flex items-center justify-between py-3 border-b border-white/[0.04] last:border-0 hover:bg-white/[0.02] px-2 rounded-lg transition-colors"
                  >
                    <div>
                      <p className="text-white/75 text-sm">{conv.title || 'Untitled'}</p>
                      <p className="text-white/25 text-xs">{new Date(conv.created_at).toLocaleString()}</p>
                    </div>
                    <span className="text-white/20 text-sm">&rarr;</span>
                  </Link>
                ))}
              </div>
            )}
          </div>

          {/* Quick launch */}
          <div className="space-y-5">
            <div className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-5">
              <p className="text-white/35 text-xs font-mono tracking-wider mb-4">// Quick Start</p>
              <div className="space-y-2">
                {agents.map((a) => (
                  <Link
                    key={a.id}
                    to={`/agents/${a.id}`}
                    className="flex items-center gap-3 px-3 py-2.5 bg-white/[0.03] hover:bg-white/[0.07] border border-white/[0.05] rounded-xl transition-colors"
                  >
                    <span className="text-xl">{a.icon || '🤖'}</span>
                    <div className="flex-1 min-w-0">
                      <p className="text-white/75 text-sm font-medium truncate">{a.name}</p>
                      <p className="text-white/25 text-xs">{a.cost} credits</p>
                    </div>
                    <span className="text-white/20 text-sm shrink-0">&rarr;</span>
                  </Link>
                ))}
              </div>
              <Link to="/agents" className="block mt-4 text-center text-primary text-sm hover:underline">
                All Agents &rarr;
              </Link>
            </div>

            <div className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-5">
              <p className="text-white/35 text-xs font-mono tracking-wider mb-4">// Account</p>
              <div className="flex items-center gap-3 mb-4">
                <div className="w-12 h-12 rounded-full bg-primary flex items-center justify-center text-base font-black text-lg">
                  {user.name.slice(0, 2).toUpperCase()}
                </div>
                <div>
                  <p className="text-white font-bold">{user.name}</p>
                  <p className="text-white/35 text-sm">{user.email}</p>
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-white/35">Status</span>
                  <span className="text-success">{user.status}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-white/35">Role</span>
                  <span className="text-white/50">{user.role}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
