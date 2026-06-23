import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { listAgents } from '../api/agent'
import type { Agent } from '../types'
import AgentCard from '../components/AgentCard'

const FEATURES = [
  { icon: '⚡', title: 'Instant Start', desc: 'No configuration needed. Pick an agent, hit go, get results in seconds.' },
  { icon: '💎', title: 'Pay Per Use', desc: 'Credit-based billing. Pay only for what you use, no subscription needed.' },
  { icon: '🔗', title: 'Multi-Engine', desc: 'Supports multiple LLM backends and workflow engines for flexible integration.' },
  { icon: '🛡️', title: 'Enterprise Security', desc: 'Encrypted data, isolated execution, full audit trail on every operation.' },
  { icon: '📈', title: 'Real-time Observability', desc: 'Usage, credits, response time — all metrics displayed in real-time.' },
  { icon: '🎯', title: 'Domain-Tuned', desc: 'Each agent is fine-tuned for its professional domain, not a generic wrapper.' },
]

const STATS = [
  { val: '10+', label: 'Professional Agents' },
  { val: '50k+', label: 'Total Calls' },
  { val: '4.7', label: 'Avg Rating' },
  { val: '< 5s', label: 'Avg Response' },
]

const agentNames = ['AI Writer', 'Contract Reviewer', 'Data Analyst', 'Smart Assistant', 'Code Expert']

export default function Landing() {
  const { user, setAuthModal } = useAuth()
  const [agents, setAgents] = useState<Agent[]>([])
  const [tick, setTick] = useState(0)

  useEffect(() => {
    listAgents({ limit: 3 }).then((r) => setAgents(r.agents)).catch(() => {})
  }, [])

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 2200)
    return () => clearInterval(id)
  }, [])

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
            <span className="text-primary text-xs font-semibold tracking-widest uppercase">AI Agent Platform</span>
          </div>

          <h1 className="text-5xl md:text-[72px] font-black text-white leading-[1.04] tracking-tight mb-4">
            Let AI Agents
            <br />
            <span
              className="bg-clip-text text-transparent"
              style={{ backgroundImage: 'linear-gradient(130deg, #00d4ff 0%, #7c3aed 100%)' }}
            >
              Be Your Superpower
            </span>
          </h1>

          <div className="h-8 overflow-hidden mb-3">
            <p key={tick} className="text-primary/70 text-lg font-mono animate-fadeInUp">
              {`> Deploying: ${agentNames[tick % agentNames.length]}`}
            </p>
          </div>

          <p className="text-white/45 text-lg max-w-2xl mx-auto mb-10 leading-relaxed">
            Professional AI agents covering writing, legal, data, customer service, development, and research. Credit-based, pay-per-use, instant start.
          </p>

          <div className="flex flex-wrap items-center justify-center gap-4">
            <Link
              to="/agents"
              className="px-8 py-4 bg-primary text-base font-black rounded-full text-base hover:bg-[#00bfe8] transition-colors shadow-[0_0_35px_rgba(0,212,255,0.3)]"
            >
              Explore Agents &rarr;
            </Link>
            {!user && (
              <button
                onClick={() => setAuthModal('register')}
                className="px-8 py-4 bg-white/[0.05] border border-white/10 text-white font-semibold rounded-full text-base hover:bg-white/10 hover:border-white/20 transition-all"
              >
                Free Register
              </button>
            )}
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="py-14 border-y border-white/[0.05] bg-white/[0.015]">
        <div className="max-w-4xl mx-auto px-6">
          <div className="flex flex-wrap justify-center gap-8 md:gap-16">
            {STATS.map((s) => (
              <div key={s.label} className="text-center">
                <div className="text-3xl md:text-4xl font-black text-white font-mono tracking-tighter">{s.val}</div>
                <div className="text-white/35 text-sm mt-1">{s.label}</div>
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
                <p className="text-primary text-xs font-mono tracking-widest uppercase mb-3">// Featured Agents</p>
                <h2 className="text-3xl md:text-4xl font-black text-white">Ready to Deploy</h2>
              </div>
              <Link to="/agents" className="text-white/35 hover:text-primary text-sm transition-colors self-start md:self-auto">
                View All &rarr;
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
            <p className="text-secondary text-xs font-mono tracking-widest uppercase mb-3">// Platform</p>
            <h2 className="text-3xl md:text-4xl font-black text-white">Why AgentHub</h2>
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
            {FEATURES.map((f) => (
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
          <p className="text-primary text-xs font-mono tracking-widest uppercase mb-4">// Get Started</p>
          <h2 className="text-4xl md:text-5xl font-black text-white mb-5">Ready?</h2>
          <p className="text-white/40 mb-8 text-base">Register now and start experiencing your first AI agent</p>
          <button
            onClick={() => setAuthModal('register')}
            className="px-10 py-4 bg-primary text-base font-black rounded-full text-lg hover:bg-[#00bfe8] transition-colors shadow-[0_0_45px_rgba(0,212,255,0.3)]"
          >
            Get Started Free &rarr;
          </button>
        </div>
      </section>

      <footer className="py-8 px-6 border-t border-white/[0.05] text-center text-white/20 text-sm">
        &copy; 2025 AgentHub - AI Agent Platform
      </footer>
    </div>
  )
}
