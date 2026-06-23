import { useState, useEffect } from 'react'
import { listAllAgents } from '../../api/agent'
import type { Agent } from '../../types'

export default function Overview() {
  const [agents, setAgents] = useState<Agent[]>([])

  useEffect(() => {
    listAllAgents({ limit: 50 }).then((r) => setAgents(r.agents)).catch(() => {})
  }, [])

  const activeCount = agents.filter((a) => a.status === 'active').length
  const totalCalls = agents.reduce((s, a) => s + (a.call_count || 0), 0)

  const stats = [
    { icon: '⬡', label: 'Total Agents', val: String(agents.length), unit: '', color: '#7c3aed' },
    { icon: '✓', label: 'Active Agents', val: String(activeCount), unit: '', color: '#10b981' },
    { icon: '⚡', label: 'Total Calls', val: totalCalls.toLocaleString(), unit: '', color: '#00d4ff' },
  ]

  return (
    <div className="p-8">
      <p className="text-secondary text-xs font-mono tracking-widest uppercase mb-2">// Overview</p>
      <h1 className="text-2xl font-black text-white mb-7">Dashboard</h1>

      <div className="grid sm:grid-cols-3 gap-4 mb-7">
        {stats.map((s) => (
          <div key={s.label} className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-5">
            <div className="text-2xl mb-3">{s.icon}</div>
            <div className="text-3xl font-black font-mono" style={{ color: s.color }}>{s.val}</div>
            <p className="text-white/30 text-sm mt-0.5">{s.label}</p>
          </div>
        ))}
      </div>

      <div className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-5">
        <p className="text-white/30 text-xs font-mono tracking-wider mb-4">// Agent Call Ranking</p>
        {[...agents].sort((a, b) => (b.call_count || 0) - (a.call_count || 0)).map((a, i) => (
          <div key={a.id} className="flex items-center gap-3 mb-3">
            <span className="text-white/20 text-sm w-4 font-mono">{i + 1}</span>
            <span className="text-base">{a.icon || '🤖'}</span>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between mb-1">
                <p className="text-white/65 text-sm truncate">{a.name}</p>
                <p className="text-white/30 text-xs font-mono ml-2">
                  {((a.call_count || 0) / 1000).toFixed(0)}k
                </p>
              </div>
              <div className="h-1 bg-white/[0.05] rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full"
                  style={{
                    width: `${((a.call_count || 0) / (agents[0]?.call_count || 1)) * 100}%`,
                    background: a.color || '#7c3aed',
                    opacity: 0.55,
                  }}
                />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
