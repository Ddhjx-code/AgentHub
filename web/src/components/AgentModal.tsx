import { useState, useEffect } from 'react'
import { useLocale } from '../contexts/LocaleContext'
import type { Agent, KnowledgeBase } from '../types'
import { listKnowledgeBases, listAgentKBs, bindAgentKB, unbindAgentKB } from '../api/knowledge'

interface Props {
  agent: Partial<Agent> | null
  onClose: () => void
  onSave: (data: Partial<Agent>) => void
}

export default function AgentModal({ agent, onClose, onSave }: Props) {
  const { t } = useLocale()
  const isNew = !agent?.id
  const [tab, setTab] = useState('basic')
  const [form, setForm] = useState({
    name: '',
    icon: '🤖',
    description: '',
    category: '',
    tags: '',
    status: 'active',
    prompt: '',
    model_name: '',
    base_url: '',
    api_key: '',
    temperature: 0.7,
    max_tokens: 2000,
    cost: 10,
    color: '#00d4ff',
  })
  const [allKBs, setAllKBs] = useState<KnowledgeBase[]>([])
  const [boundKBIds, setBoundKBIds] = useState<Set<number>>(new Set())

  useEffect(() => {
    if (agent) {
      setForm({
        name: agent.name || '',
        icon: agent.icon || '🤖',
        description: agent.short_desc || agent.description || '',
        category: agent.category || '',
        tags: (agent.tags || []).join(', '),
        status: agent.status || 'active',
        prompt: agent.prompt || '',
        model_name: agent.model_name || '',
        base_url: agent.base_url || '',
        api_key: agent.api_key || '',
        temperature: agent.temperature ?? 0.7,
        max_tokens: agent.max_tokens || 2000,
        cost: agent.cost || 10,
        color: agent.color || '#00d4ff',
      })
    }
    listKnowledgeBases().then(setAllKBs).catch(() => {})
    if (agent?.id) {
      listAgentKBs(agent.id).then((kbs) => setBoundKBIds(new Set(kbs.map((k) => k.id)))).catch(() => {})
    }
  }, [agent])

  if (!agent) return null

  const set = (k: string, v: string | number) => setForm((f) => ({ ...f, [k]: v }))

  const handleSave = () => {
    const { description, ...rest } = form
    onSave({
      ...rest,
      short_desc: description,
      tags: form.tags.split(',').map((s) => s.trim()).filter(Boolean),
    } as unknown as Partial<Agent>)
  }

  const toggleKB = async (kbId: number) => {
    if (!agent.id) return
    try {
      if (boundKBIds.has(kbId)) {
        await unbindAgentKB(agent.id, kbId)
        setBoundKBIds((prev) => { const n = new Set(prev); n.delete(kbId); return n })
      } else {
        await bindAgentKB(agent.id, kbId)
        setBoundKBIds((prev) => new Set(prev).add(kbId))
      }
    } catch {
      /* ignore */
    }
  }

  const tabs: [string, string][] = [
    ['basic', t.agentModal.tabBasic],
    ['prompt', t.agentModal.tabPrompt],
    ['llm', t.agentModal.tabLLM],
    ['kb', t.agentModal.tabKB],
  ]
  const inp = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-3 py-2.5 text-white/80 text-sm focus:outline-none focus:border-secondary/50 transition-colors placeholder-white/20'
  const lbl = 'text-white/40 text-xs mb-1.5 block'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={onClose}>
      <div
        className="relative w-full max-w-2xl mx-4 bg-[#090c1c] border border-white/10 rounded-3xl shadow-[0_0_60px_rgba(124,58,237,0.15)] max-h-[90vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-7 pt-6 pb-0">
          <div>
            <h2 className="text-white font-black text-lg">{isNew ? t.agentModal.newTitle : form.name}</h2>
            <p className="text-white/30 text-sm">{isNew ? t.agentModal.newDesc : t.agentModal.editDesc}</p>
          </div>
          <button className="w-8 h-8 flex items-center justify-center rounded-full bg-white/5 text-white/40 hover:text-white/80 text-lg" onClick={onClose}>
            x
          </button>
        </div>

        <div className="flex gap-1 px-7 pt-5 pb-0">
          {tabs.map(([k, l]) => (
            <button
              key={k}
              onClick={() => setTab(k)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                tab === k ? 'bg-secondary text-white' : 'text-white/35 hover:text-white/65 hover:bg-white/5'
              }`}
            >
              {l}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-7 pt-5 space-y-4">
          {tab === 'basic' && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <div><label className={lbl}>{t.agentModal.name}</label><input className={inp} value={form.name} onChange={(e) => set('name', e.target.value)} /></div>
                <div><label className={lbl}>{t.agentModal.icon}</label><input className={inp} value={form.icon} onChange={(e) => set('icon', e.target.value)} /></div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div><label className={lbl}>{t.agentModal.category}</label><input className={inp} value={form.category} onChange={(e) => set('category', e.target.value)} /></div>
                <div><label className={lbl}>{t.agentModal.costLabel}</label><input type="number" className={inp} value={form.cost} onChange={(e) => set('cost', +e.target.value)} /></div>
              </div>
              <div><label className={lbl}>{t.agentModal.description}</label><textarea className={`${inp} resize-none`} rows={3} value={form.description} onChange={(e) => set('description', e.target.value)} /></div>
              <div className="grid grid-cols-2 gap-4">
                <div><label className={lbl}>{t.agentModal.tags}</label><input className={inp} value={form.tags} onChange={(e) => set('tags', e.target.value)} /></div>
                <div><label className={lbl}>{t.agentModal.color}</label><input className={inp} value={form.color} onChange={(e) => set('color', e.target.value)} /></div>
              </div>
              <div className="flex items-center justify-between p-3 bg-white/[0.03] rounded-xl border border-white/[0.06]">
                <span className="text-white/55 text-sm">{t.agentModal.statusLabel}</span>
                <div className="flex gap-2">
                  {(['active', 'inactive'] as const).map((s) => (
                    <button
                      key={s}
                      onClick={() => set('status', s)}
                      className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                        form.status === s
                          ? s === 'active'
                            ? 'bg-success/20 text-success border border-success/25'
                            : 'bg-danger/15 text-danger border border-danger/20'
                          : 'bg-white/[0.04] text-white/30 border border-white/[0.07]'
                      }`}
                    >
                      {s === 'active' ? t.admin.active : t.admin.inactive}
                    </button>
                  ))}
                </div>
              </div>
            </>
          )}

          {tab === 'prompt' && (
            <>
              <div>
                <label className={lbl}>{t.agentModal.prompt}</label>
                <textarea className={`${inp} resize-none font-mono text-xs`} rows={8} value={form.prompt} onChange={(e) => set('prompt', e.target.value)} placeholder={t.agentModal.promptPlaceholder} />
              </div>
            </>
          )}

          {tab === 'llm' && (
            <>
              <div><label className={lbl}>{t.agentModal.modelName}</label><input className={`${inp} font-mono`} value={form.model_name} onChange={(e) => set('model_name', e.target.value)} placeholder="e.g. gpt-4o" /></div>
              <div><label className={lbl}>{t.agentModal.baseURL}</label><input className={`${inp} font-mono text-xs`} value={form.base_url} onChange={(e) => set('base_url', e.target.value)} placeholder="https://api.openai.com" /></div>
              <div><label className={lbl}>{t.agentModal.apiKey}</label><input type="password" className={`${inp} font-mono`} value={form.api_key} onChange={(e) => set('api_key', e.target.value)} placeholder="sk-..." /></div>
              <div>
                <label className={lbl}>{t.agentModal.tempLabel}: {form.temperature}</label>
                <input type="range" min="0" max="1" step="0.05" value={form.temperature} onChange={(e) => set('temperature', +e.target.value)} className="w-full accent-secondary" />
                <div className="flex justify-between text-white/20 text-xs mt-1"><span>{t.agentModal.precise}</span><span>{t.agentModal.creative}</span></div>
              </div>
              <div><label className={lbl}>{t.agentModal.maxTokens}</label><input type="number" className={inp} value={form.max_tokens} onChange={(e) => set('max_tokens', +e.target.value)} /></div>
            </>
          )}

          {tab === 'kb' && (
            <>
              {!agent.id ? (
                <p className="text-white/30 text-sm py-4">{t.agentModal.saveFirst}</p>
              ) : allKBs.length === 0 ? (
                <p className="text-white/30 text-sm py-4">{t.agentModal.noKBs}</p>
              ) : (
                <div className="space-y-2">
                  {allKBs.map((kb) => (
                    <div
                      key={kb.id}
                      className={`flex items-center justify-between p-3 rounded-xl border transition-all cursor-pointer ${
                        boundKBIds.has(kb.id)
                          ? 'bg-secondary/15 border-secondary/25'
                          : 'bg-white/[0.03] border-white/[0.07] hover:border-white/15'
                      }`}
                      onClick={() => toggleKB(kb.id)}
                    >
                      <div>
                        <p className="text-white/75 text-sm font-medium">{kb.name}</p>
                        <p className="text-white/30 text-xs">{kb.description || t.agentModal.noDesc}</p>
                      </div>
                      <span className={`text-xs font-semibold ${boundKBIds.has(kb.id) ? 'text-secondary' : 'text-white/25'}`}>
                        {boundKBIds.has(kb.id) ? t.agentModal.bound : t.agentModal.unbound}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>

        <div className="flex justify-end gap-3 px-7 py-5 border-t border-white/[0.06]">
          <button className="px-5 py-2.5 bg-white/[0.04] border border-white/[0.08] text-white/50 rounded-xl text-sm hover:bg-white/[0.08] transition-colors" onClick={onClose}>
            {t.agentModal.cancel}
          </button>
          <button className="px-6 py-2.5 bg-secondary text-white font-bold rounded-xl text-sm hover:bg-[#6d28d9] transition-colors shadow-[0_0_20px_rgba(124,58,237,0.3)]" onClick={handleSave}>
            {t.agentModal.save}
          </button>
        </div>
      </div>
    </div>
  )
}
