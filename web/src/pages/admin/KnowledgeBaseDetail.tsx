import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getKnowledgeBase, listDocuments, uploadDocument, deleteDocument } from '../../api/knowledge'
import type { KnowledgeBase, Document } from '../../types'
import { useAuth } from '../../contexts/AuthContext'

export default function KnowledgeBaseDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { flash } = useAuth()

  const [kb, setKB] = useState<KnowledgeBase | null>(null)
  const [docs, setDocs] = useState<Document[]>([])
  const [showUpload, setShowUpload] = useState(false)
  const [docName, setDocName] = useState('')
  const [docContent, setDocContent] = useState('')
  const [uploading, setUploading] = useState(false)

  const load = () => {
    if (!id) return
    getKnowledgeBase(Number(id)).then(setKB).catch(() => navigate('/admin/knowledge-bases'))
    listDocuments(Number(id)).then(setDocs).catch(() => {})
  }

  useEffect(() => { load() }, [id])

  const handleUpload = async () => {
    if (!id || !docName.trim() || !docContent.trim()) return
    setUploading(true)
    try {
      await uploadDocument(Number(id), docName.trim(), docContent.trim())
      flash('Document uploaded')
      setShowUpload(false)
      setDocName('')
      setDocContent('')
      load()
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Upload failed', 'error')
    } finally {
      setUploading(false)
    }
  }

  const handleDeleteDoc = async (docId: number) => {
    if (!id || !confirm('Delete this document?')) return
    try {
      await deleteDocument(Number(id), docId)
      flash('Document deleted')
      load()
    } catch {
      flash('Delete failed', 'error')
    }
  }

  if (!kb) {
    return <div className="p-8 text-white/30">Loading...</div>
  }

  const inp = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-3 py-2.5 text-white/80 text-sm focus:outline-none focus:border-secondary/50 transition-colors placeholder-white/20'
  const TH = 'text-white/30 text-xs font-semibold uppercase tracking-wider py-3 px-4 text-left'
  const TD = 'py-3 px-4 text-sm'

  return (
    <div className="p-8">
      <button onClick={() => navigate('/admin/knowledge-bases')} className="text-white/35 hover:text-white/75 text-sm mb-6 transition-colors">
        &larr; Back to Knowledge Bases
      </button>

      {/* KB info */}
      <div className="bg-white/[0.03] border border-white/[0.07] rounded-2xl p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-black text-white mb-1">{kb.name}</h1>
            {kb.description && <p className="text-white/45 text-sm mb-4">{kb.description}</p>}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {[
                { label: 'Model', val: kb.embedding_model || '-' },
                { label: 'Dimension', val: String(kb.dimension || '-') },
                { label: 'Chunk Size', val: String(kb.chunk_size || 512) },
                { label: 'Overlap', val: String(kb.chunk_overlap || 64) },
              ].map((s) => (
                <div key={s.label}>
                  <p className="text-white/30 text-xs">{s.label}</p>
                  <p className="text-white/65 font-mono text-sm">{s.val}</p>
                </div>
              ))}
            </div>
          </div>
          <span className={`px-2.5 py-1 rounded-full text-xs font-semibold ${
            kb.status === 'active' ? 'bg-success/15 text-success' : 'bg-white/[0.05] text-white/30'
          }`}>
            {kb.status}
          </span>
        </div>
      </div>

      {/* Documents */}
      <div className="flex items-center justify-between mb-4">
        <p className="text-secondary text-xs font-mono tracking-widest uppercase">// Documents</p>
        <button
          onClick={() => setShowUpload(true)}
          className="px-4 py-2 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors"
        >
          + Upload Document
        </button>
      </div>

      <div className="bg-white/[0.02] border border-white/[0.07] rounded-2xl overflow-hidden">
        <table className="w-full">
          <thead className="border-b border-white/[0.06]">
            <tr>
              <th className={TH}>Name</th>
              <th className={TH}>Chunks</th>
              <th className={TH}>Status</th>
              <th className={TH}>Created</th>
              <th className={TH}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {docs.map((doc, i) => (
              <tr key={doc.id} className={`border-b border-white/[0.04] last:border-0 ${i % 2 !== 0 ? 'bg-white/[0.01]' : ''}`}>
                <td className={TD}><span className="text-white/75">{doc.name}</span></td>
                <td className={TD}><span className="text-white/40 font-mono">{doc.chunk_count}</span></td>
                <td className={TD}>
                  <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
                    doc.status === 'completed' ? 'bg-success/15 text-success'
                      : doc.status === 'failed' ? 'bg-danger/15 text-danger'
                        : 'bg-accent/15 text-accent'
                  }`}>
                    {doc.status}
                  </span>
                </td>
                <td className={TD}><span className="text-white/30 text-xs">{new Date(doc.created_at).toLocaleString()}</span></td>
                <td className={TD}>
                  <button onClick={() => handleDeleteDoc(doc.id)} className="px-2.5 py-1 bg-white/[0.04] text-white/30 rounded-lg text-xs hover:bg-danger/15 hover:text-danger transition-colors">
                    Delete
                  </button>
                </td>
              </tr>
            ))}
            {docs.length === 0 && (
              <tr><td colSpan={5} className="text-center py-8 text-white/25 text-sm">No documents</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Upload Modal */}
      {showUpload && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={() => setShowUpload(false)}>
          <div className="relative w-full max-w-lg mx-4 bg-[#090c1c] border border-white/10 rounded-3xl p-7 shadow-[0_0_60px_rgba(124,58,237,0.15)]" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-white font-black text-lg mb-5">Upload Document</h2>
            <div className="space-y-4">
              <div>
                <label className="text-white/40 text-xs mb-1.5 block">Document Name</label>
                <input className={inp} value={docName} onChange={(e) => setDocName(e.target.value)} placeholder="e.g. guide.txt" />
              </div>
              <div>
                <label className="text-white/40 text-xs mb-1.5 block">Content (plain text)</label>
                <textarea className={`${inp} resize-none font-mono text-xs`} rows={10} value={docContent} onChange={(e) => setDocContent(e.target.value)} placeholder="Paste document content here..." />
              </div>
            </div>
            <div className="flex justify-end gap-3 mt-6">
              <button className="px-5 py-2.5 bg-white/[0.04] border border-white/[0.08] text-white/50 rounded-xl text-sm hover:bg-white/[0.08] transition-colors" onClick={() => setShowUpload(false)}>Cancel</button>
              <button
                className="px-6 py-2.5 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors shadow-[0_0_20px_rgba(124,58,237,0.3)] disabled:opacity-50"
                onClick={handleUpload}
                disabled={uploading || !docName.trim() || !docContent.trim()}
              >
                {uploading ? 'Uploading...' : 'Upload'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
