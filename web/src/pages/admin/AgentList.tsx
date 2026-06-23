import { useState, useEffect } from 'react'
import { listAllAgents, createAgent, updateAgent, deleteAgent, toggleAgent } from '../../api/agent'
import type { Agent } from '../../types'
import AgentModal from '../../components/AgentModal'
import { useAuth } from '../../contexts/AuthContext'

export default function AgentList() {
  const { flash } = useAuth()
  const [agents, setAgents] = useState<Agent[]>([])
  const [modal, setModal] = useState<Partial<Agent> | null>(null)
  const [loading, setLoading] = useState(true)

  const loadAgents = () => {
    setLoading(true)
    listAllAgents({ limit: 100 })
      .then((r) => setAgents(r.agents))
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadAgents() }, [])

  const handleSave = async (data: Partial<Agent>) => {
    try {
      if (modal?.id) {
        await updateAgent(modal.id, data)
        flash('Agent updated')
      } else {
        await createAgent(data)
        flash('Agent created')
      }
      setModal(null)
      loadAgents()
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Save failed', 'error')
    }
  }

  const handleToggle = async (id: number) => {
    try {
      await toggleAgent(id)
      loadAgents()
    } catch {
      flash('Toggle failed', 'error')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this agent?')) return
    try {
      await deleteAgent(id)
      flash('Agent deleted')
      loadAgents()
    } catch {
      flash('Delete failed', 'error')
    }
  }

  const TH = 'text-white/30 text-xs font-semibold uppercase tracking-wider py-3 px-4 text-left'
  const TD = 'py-3 px-4 text-sm'

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <p className="text-secondary text-xs font-mono tracking-widest uppercase mb-1">// Agent Management</p>
          <h1 className="text-2xl font-black text-white">Agent List</h1>
        </div>
        <button
          onClick={() => setModal({})}
          className="px-5 py-2.5 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors shadow-[0_0_15px_rgba(124,58,237,0.25)]"
        >
          + New Agent
        </button>
      </div>

      {loading ? (
        <p className="text-white/30 text-sm py-8">Loading...</p>
      ) : (
        <div className="bg-white/[0.02] border border-white/[0.07] rounded-2xl overflow-hidden">
          <table className="w-full">
            <thead className="border-b border-white/[0.06]">
              <tr>
                <th className={TH}>Agent</th>
                <th className={TH}>Category</th>
                <th className={TH}>Model</th>
                <th className={TH}>Cost</th>
                <th className={TH}>Calls</th>
                <th className={TH}>Status</th>
                <th className={TH}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {agents.map((a, i) => (
                <tr key={a.id} className={`border-b border-white/[0.04] last:border-0 ${i % 2 !== 0 ? 'bg-white/[0.01]' : ''}`}>
                  <td className={TD}>
                    <div className="flex items-center gap-2.5">
                      <div
                        className="w-8 h-8 rounded-lg flex items-center justify-center text-base shrink-0"
                        style={{ background: `${a.color || '#00d4ff'}18`, border: `1px solid ${a.color || '#00d4ff'}25` }}
                      >
                        {a.icon || '🤖'}
                      </div>
                      <span className="text-white/75 font-medium">{a.name}</span>
                    </div>
                  </td>
                  <td className={TD}><span className="text-white/40">{a.category || '-'}</span></td>
                  <td className={TD}><span className="text-white/40 font-mono text-xs">{a.model_name || '-'}</span></td>
                  <td className={TD}>
                    <span className="text-accent font-mono font-bold">{a.cost}</span>
                    <span className="text-white/25 text-xs ml-1">cr</span>
                  </td>
                  <td className={TD}><span className="text-white/40 font-mono">{((a.call_count || 0) / 1000).toFixed(1)}k</span></td>
                  <td className={TD}>
                    <button
                      onClick={() => handleToggle(a.id)}
                      className={`px-2.5 py-1 rounded-full text-xs font-semibold transition-all ${
                        a.status === 'active'
                          ? 'bg-success/15 text-success border border-success/25'
                          : 'bg-white/[0.05] text-white/30 border border-white/[0.08]'
                      }`}
                    >
                      {a.status === 'active' ? 'Active' : 'Inactive'}
                    </button>
                  </td>
                  <td className={TD}>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setModal(a)}
                        className="px-2.5 py-1 bg-secondary/15 text-secondary rounded-lg text-xs hover:bg-secondary/25 transition-colors border border-secondary/20"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(a.id)}
                        className="px-2.5 py-1 bg-white/[0.04] text-white/30 rounded-lg text-xs hover:bg-danger/15 hover:text-danger transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <AgentModal agent={modal} onClose={() => setModal(null)} onSave={handleSave} />
    </div>
  )
}
