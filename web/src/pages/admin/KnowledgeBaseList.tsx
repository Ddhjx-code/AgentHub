import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { listKnowledgeBases, createKnowledgeBase, updateKnowledgeBase, deleteKnowledgeBase } from '../../api/knowledge'
import type { KnowledgeBase } from '../../types'
import { useAuth } from '../../contexts/AuthContext'

export default function KnowledgeBaseList() {
  const navigate = useNavigate()
  const { flash } = useAuth()
  const [kbs, setKBs] = useState<KnowledgeBase[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [editKB, setEditKB] = useState<KnowledgeBase | null>(null)
  const [form, setForm] = useState({
    name: '',
    description: '',
    embedding_base_url: '',
    embedding_api_key: '',
    embedding_model: '',
    chunk_size: 512,
    chunk_overlap: 64,
  })

  const loadKBs = () => {
    listKnowledgeBases().then(setKBs).catch(() => {})
  }

  useEffect(() => { loadKBs() }, [])

  const resetForm = () => {
    setForm({ name: '', description: '', embedding_base_url: '', embedding_api_key: '', embedding_model: '', chunk_size: 512, chunk_overlap: 64 })
  }

  const openEdit = (kb: KnowledgeBase) => {
    setEditKB(kb)
    setForm({
      name: kb.name,
      description: kb.description,
      embedding_base_url: kb.embedding_base_url,
      embedding_api_key: kb.embedding_api_key,
      embedding_model: kb.embedding_model,
      chunk_size: kb.chunk_size || 512,
      chunk_overlap: kb.chunk_overlap || 64,
    })
    setShowCreate(true)
  }

  const handleSave = async () => {
    try {
      if (editKB) {
        await updateKnowledgeBase(editKB.id, form)
        flash('Knowledge base updated')
      } else {
        await createKnowledgeBase(form)
        flash('Knowledge base created')
      }
      setShowCreate(false)
      setEditKB(null)
      resetForm()
      loadKBs()
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Save failed', 'error')
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this knowledge base?')) return
    try {
      await deleteKnowledgeBase(id)
      flash('Knowledge base deleted')
      loadKBs()
    } catch {
      flash('Delete failed', 'error')
    }
  }

  const inp = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-3 py-2.5 text-white/80 text-sm focus:outline-none focus:border-secondary/50 transition-colors placeholder-white/20'
  const lbl = 'text-white/40 text-xs mb-1.5 block'
  const TH = 'text-white/30 text-xs font-semibold uppercase tracking-wider py-3 px-4 text-left'
  const TD = 'py-3 px-4 text-sm'

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <p className="text-secondary text-xs font-mono tracking-widest uppercase mb-1">// Knowledge Base</p>
          <h1 className="text-2xl font-black text-white">Knowledge Bases</h1>
        </div>
        <button
          onClick={() => { resetForm(); setEditKB(null); setShowCreate(true) }}
          className="px-5 py-2.5 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors shadow-[0_0_15px_rgba(124,58,237,0.25)]"
        >
          + New KB
        </button>
      </div>

      <div className="bg-white/[0.02] border border-white/[0.07] rounded-2xl overflow-hidden">
        <table className="w-full">
          <thead className="border-b border-white/[0.06]">
            <tr>
              <th className={TH}>Name</th>
              <th className={TH}>Model</th>
              <th className={TH}>Chunk Size</th>
              <th className={TH}>Status</th>
              <th className={TH}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {kbs.map((kb, i) => (
              <tr key={kb.id} className={`border-b border-white/[0.04] last:border-0 ${i % 2 !== 0 ? 'bg-white/[0.01]' : ''}`}>
                <td className={TD}>
                  <button onClick={() => navigate(`/admin/knowledge-bases/${kb.id}`)} className="text-white/75 font-medium hover:text-primary transition-colors text-left">
                    {kb.name}
                  </button>
                  {kb.description && <p className="text-white/30 text-xs mt-0.5">{kb.description}</p>}
                </td>
                <td className={TD}><span className="text-white/40 font-mono text-xs">{kb.embedding_model || '-'}</span></td>
                <td className={TD}><span className="text-white/40 font-mono">{kb.chunk_size || 512}</span></td>
                <td className={TD}>
                  <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
                    kb.status === 'active' ? 'bg-success/15 text-success' : 'bg-white/[0.05] text-white/30'
                  }`}>
                    {kb.status}
                  </span>
                </td>
                <td className={TD}>
                  <div className="flex items-center gap-2">
                    <button onClick={() => navigate(`/admin/knowledge-bases/${kb.id}`)} className="px-2.5 py-1 bg-secondary/15 text-secondary rounded-lg text-xs hover:bg-secondary/25 transition-colors border border-secondary/20">Detail</button>
                    <button onClick={() => openEdit(kb)} className="px-2.5 py-1 bg-white/[0.04] text-white/40 rounded-lg text-xs hover:bg-white/[0.08] transition-colors">Edit</button>
                    <button onClick={() => handleDelete(kb.id)} className="px-2.5 py-1 bg-white/[0.04] text-white/30 rounded-lg text-xs hover:bg-danger/15 hover:text-danger transition-colors">Delete</button>
                  </div>
                </td>
              </tr>
            ))}
            {kbs.length === 0 && (
              <tr><td colSpan={5} className="text-center py-8 text-white/25 text-sm">No knowledge bases</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Create/Edit Modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={() => { setShowCreate(false); setEditKB(null) }}>
          <div className="relative w-full max-w-lg mx-4 bg-[#090c1c] border border-white/10 rounded-3xl p-7 shadow-[0_0_60px_rgba(124,58,237,0.15)]" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-white font-black text-lg mb-5">{editKB ? 'Edit Knowledge Base' : 'New Knowledge Base'}</h2>
            <div className="space-y-4">
              <div><label className={lbl}>Name</label><input className={inp} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></div>
              <div><label className={lbl}>Description</label><textarea className={`${inp} resize-none`} rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></div>
              <div><label className={lbl}>Embedding Base URL</label><input className={`${inp} font-mono text-xs`} value={form.embedding_base_url} onChange={(e) => setForm({ ...form, embedding_base_url: e.target.value })} placeholder="http://localhost:11434" /></div>
              <div><label className={lbl}>Embedding API Key</label><input type="password" className={`${inp} font-mono`} value={form.embedding_api_key} onChange={(e) => setForm({ ...form, embedding_api_key: e.target.value })} /></div>
              <div><label className={lbl}>Embedding Model</label><input className={`${inp} font-mono`} value={form.embedding_model} onChange={(e) => setForm({ ...form, embedding_model: e.target.value })} placeholder="nomic-embed-text" /></div>
              <div className="grid grid-cols-2 gap-4">
                <div><label className={lbl}>Chunk Size</label><input type="number" className={inp} value={form.chunk_size} onChange={(e) => setForm({ ...form, chunk_size: +e.target.value })} /></div>
                <div><label className={lbl}>Chunk Overlap</label><input type="number" className={inp} value={form.chunk_overlap} onChange={(e) => setForm({ ...form, chunk_overlap: +e.target.value })} /></div>
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button className="px-5 py-2.5 bg-white/[0.04] border border-white/[0.08] text-white/50 rounded-xl text-sm hover:bg-white/[0.08] transition-colors" onClick={() => { setShowCreate(false); setEditKB(null) }}>Cancel</button>
              <button className="px-6 py-2.5 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors shadow-[0_0_20px_rgba(124,58,237,0.3)]" onClick={handleSave}>Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
