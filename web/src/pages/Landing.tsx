import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useLocale } from '../contexts/LocaleContext'
import { listAgents } from '../api/agent'
import type { Agent } from '../types'
import AgentCard from '../components/AgentCard'

const STATS_VALS = ['10+', '50k+', '4.7', '< 5s']

export default function Landing() {
  const { user, setAuthModal } = useAuth()
  const { t } = useLocale()
  const [agents, setAgents] = useState<Agent[]>([])
  const [tick, setTick] = useState(0)

  useEffect(() => {
    listAgents({ limit: 3 }).then((r) => setAgents(r.agents)).catch(() => {})
  }, [])

  useEffect(() => {
    const id = setInterval(() => setTick((prev) => prev + 1), 2200)
    return () => clearInterval(id)
  }, [])

  const statsLabels = [t.landing.stats.agents, t.landing.stats.calls, t.landing.stats.rating, t.landing.stats.response]
  const agentNames = t.landing.agentNames

  return (
    <div>
      {/* Hero */}
      <section className="relative min-h-[100svh] flex flex-col items-center justify-center px-6 pb-24 overflow-hidden">
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] rounded-full bg-primary/[0.04] blur-[130px]" />
          <div className="absolute top-1/4 right-1/4 w-[300px] h-[300px] rounded-full bg-secondary/[0.05] blur-[100px]" />
        </div>

        <div className="relative z-10 text-center max-w-4xl mx-auto">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 mb-8 bg-primary/10 border border-primary/20 rounded-full">
            <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
            <span className="text-primary text-xs font-semibold tracking-widest uppercase">{t.landing.badge}</span>
          </div>

          <h1 className="text-5xl md:text-[72px] font-black text-white leading-[1.04] tracking-tight mb-4">
            {t.landing.title1}
            <br />
            <span
              className="bg-clip-text text-transparent"
              style={{ backgroundImage: 'linear-gradient(130deg, #00d4ff 0%, #7c3aed 100%)' }}
            >
              {t.landing.title2}
            </span>
          </h1>

          <div className="h-8 overflow-hidden mb-3">
            <p key={tick} className="text-primary/70 text-lg font-mono animate-fadeInUp">
              {t.landing.deploying.replace('{name}', agentNames[tick % agentNames.length])}
            </p>
          </div>

          <p className="text-white/45 text-lg max-w-2xl mx-auto mb-10 leading-relaxed">
            {t.landing.subtitle}
          </p>

          <div className="flex flex-wrap items-center justify-center gap-4">
            <Link
              to="/agents"
              className="px-8 py-4 bg-primary text-base font-black rounded-full text-base hover:bg-[#00bfe8] transition-colors shadow-[0_0_35px_rgba(0,212,255,0.3)]"
            >
              {t.landing.explore} &rarr;
            </Link>
            {!user && (
              <button
                onClick={() => setAuthModal('register')}
                className="px-8 py-4 bg-white/[0.05] border border-white/10 text-white font-semibold rounded-full text-base hover:bg-white/10 hover:border-white/20 transition-all"
              >
                {t.landing.freeRegister}
              </button>
            )}
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="py-14 border-y border-white/[0.05] bg-white/[0.015]">
        <div className="max-w-4xl mx-auto px-6">
          <div className="flex flex-wrap justify-center gap-8 md:gap-16">
            {STATS_VALS.map((val, i) => (
              <div key={statsLabels[i]} className="text-center">
                <div className="text-3xl md:text-4xl font-black text-white font-mono tracking-tighter">{val}</div>
                <div className="text-white/35 text-sm mt-1">{statsLabels[i]}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Featured Agents */}
      {agents.length > 0 && (
        <section className="py-24 px-6">
          <div className="max-w-6xl mx-auto">
            <div className="mb-12 flex flex-col md:flex-row md:items-end md:justify-between gap-4">
              <div>
                <p className="text-primary text-xs font-mono tracking-widest uppercase mb-3">// {t.landing.featured}</p>
                <h2 className="text-3xl md:text-4xl font-black text-white">{t.landing.readyToDeploy}</h2>
              </div>
              <Link to="/agents" className="text-white/35 hover:text-primary text-sm transition-colors self-start md:self-auto">
                {t.landing.viewAll} &rarr;
              </Link>
            </div>
            <div className="grid md:grid-cols-3 gap-5">
              {agents.map((agent) => (
                <AgentCard key={agent.id} agent={agent} />
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Features */}
      <section className="py-24 px-6 bg-white/[0.015] border-t border-white/[0.05]">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <p className="text-secondary text-xs font-mono tracking-widest uppercase mb-3">// {t.landing.platformLabel}</p>
            <h2 className="text-3xl md:text-4xl font-black text-white">{t.landing.whyTitle}</h2>
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
            {t.landing.features.map((f) => (
              <div
                key={f.title}
                className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-6 hover:border-white/15 transition-all"
              >
                <div className="text-3xl mb-4">{f.icon}</div>
                <h3 className="font-bold text-white mb-2">{f.title}</h3>
                <p className="text-white/45 text-sm leading-relaxed">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-32 px-6 text-center">
        <div className="relative max-w-xl mx-auto">
          <div className="absolute inset-0 bg-primary/[0.04] blur-[80px] rounded-full pointer-events-none" />
          <p className="text-primary text-xs font-mono tracking-widest uppercase mb-4">// {t.landing.ctaLabel}</p>
          <h2 className="text-4xl md:text-5xl font-black text-white mb-5">{t.landing.ctaTitle}</h2>
          <p className="text-white/40 mb-8 text-base">{t.landing.ctaDesc}</p>
          <button
            onClick={() => setAuthModal('register')}
            className="px-10 py-4 bg-primary text-base font-black rounded-full text-lg hover:bg-[#00bfe8] transition-colors shadow-[0_0_45px_rgba(0,212,255,0.3)]"
          >
            {t.landing.ctaBtn} &rarr;
          </button>
        </div>
      </section>

      <footer className="py-8 px-6 border-t border-white/[0.05] text-center text-white/20 text-sm">
        {t.landing.footer}
      </footer>
    </div>
  )
}
