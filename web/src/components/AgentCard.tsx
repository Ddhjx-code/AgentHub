import { useNavigate } from 'react-router-dom'
import type { Agent } from '../types'

interface Props {
  agent: Agent
}

export default function AgentCard({ agent }: Props) {
  const navigate = useNavigate()
  const color = agent.color || '#00d4ff'
  const tags = agent.tags || []

  return (
    <div
      onClick={() => navigate(`/agents/${agent.id}`)}
      className="group relative overflow-hidden bg-white/[0.03] border border-white/[0.07] rounded-2xl p-6 cursor-pointer hover:border-white/[0.18] hover:bg-white/[0.05] transition-all duration-200 hover:shadow-[0_0_28px_rgba(0,212,255,0.07)]"
    >
      <div className="absolute inset-0 pointer-events-none overflow-hidden rounded-2xl">
        <div className="absolute -top-full left-0 right-0 h-full bg-gradient-to-b from-transparent via-primary/[0.07] to-transparent group-hover:translate-y-[200%] transition-transform duration-700 ease-in-out" />
      </div>

      <div className="flex items-start gap-4 mb-4">
        <div
          className="w-12 h-12 rounded-xl flex items-center justify-center text-2xl shrink-0"
          style={{ background: `${color}18`, border: `1px solid ${color}30` }}
        >
          {agent.icon || '🤖'}
        </div>
        <div className="min-w-0">
          <h3 className="font-bold text-white">{agent.name}</h3>
          <p className="text-white/45 text-sm mt-0.5 line-clamp-2">{agent.description}</p>
        </div>
      </div>

      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-5">
          {tags.map((t) => (
            <span
              key={t}
              className="px-2 py-0.5 bg-white/[0.04] border border-white/[0.07] rounded-full text-white/35 text-[11px]"
            >
              {t}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between pt-4 border-t border-white/[0.06]">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <span className="text-accent">&#9670;</span>
            <span className="text-accent font-mono font-bold text-sm">{agent.cost}</span>
            <span className="text-white/30 text-xs">credits</span>
          </div>
          {agent.category && <span className="text-white/25 text-xs">{agent.category}</span>}
        </div>
        <span className="text-white/30 text-xs font-mono">{(agent.call_count / 1000).toFixed(1)}k calls</span>
      </div>
    </div>
  )
}
