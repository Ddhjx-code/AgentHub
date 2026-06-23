import { useState, useEffect } from 'react'
import { useLocale } from '../contexts/LocaleContext'
import { listAgents } from '../api/agent'
import type { Agent } from '../types'
import AgentCard from '../components/AgentCard'

export default function AgentMarket() {
  const { t } = useLocale()
  const [agents, setAgents] = useState<Agent[]>([])
  const [category, setCategory] = useState(t.market.categories[0])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)

  const categories = t.market.categories

  useEffect(() => {
    setLoading(true)
    const params: Record<string, string | number> = { limit: 50 }
    if (category !== categories[0]) params.category = category
    if (query) params.tag = query
    listAgents(params)
      .then((r) => setAgents(r.agents))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [category, query])

  return (
    <div className="min-h-screen py-14 px-6">
      <div className="max-w-6xl mx-auto">
        <div className="mb-10">
          <p className="text-primary text-xs font-mono tracking-widest uppercase mb-3">// {t.market.label}</p>
          <h1 className="text-4xl font-black text-white mb-3">{t.market.title}</h1>
          <p className="text-white/40">{t.market.subtitle}</p>
        </div>

        <div className="mb-6">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t.market.search}
            className="w-full md:w-80 bg-white/[0.04] border border-white/[0.08] rounded-xl px-4 py-3 text-white placeholder-white/25 text-sm focus:outline-none focus:border-primary/40 transition-colors"
          />
        </div>

        <div className="flex flex-wrap gap-2 mb-9">
          {categories.map((c) => (
            <button
              key={c}
              onClick={() => setCategory(c)}
              className={`px-4 py-2 rounded-full text-sm font-medium transition-all ${
                category === c
                  ? 'bg-primary text-base shadow-[0_0_15px_rgba(0,212,255,0.2)]'
                  : 'bg-white/[0.04] border border-white/[0.08] text-white/45 hover:text-white/80 hover:border-white/20'
              }`}
            >
              {c}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="text-center py-24">
            <p className="text-white/30">{t.market.loading}</p>
          </div>
        ) : agents.length > 0 ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
            {agents.map((a) => (
              <AgentCard key={a.id} agent={a} />
            ))}
          </div>
        ) : (
          <div className="text-center py-24">
            <div className="text-5xl mb-4">🔭</div>
            <p className="text-white/30">{t.market.noResults}</p>
          </div>
        )}
      </div>
    </div>
  )
}
