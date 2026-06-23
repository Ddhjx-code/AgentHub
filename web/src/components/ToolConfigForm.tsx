import { useLocale } from '../contexts/LocaleContext'

export interface ToolFormData {
  _id: string
  name: string
  type: string
  description: string
  input_schema: string
  config: Record<string, unknown>
}

interface Props {
  tool: ToolFormData
  onChange: (tool: ToolFormData) => void
  onDelete: () => void
}

const inp = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-3 py-2.5 text-white/80 text-sm focus:outline-none focus:border-secondary/50 transition-colors placeholder-white/20'
const lbl = 'text-white/40 text-xs mb-1.5 block'

export default function ToolConfigForm({ tool, onChange, onDelete }: Props) {
  const { t } = useLocale()

  const set = (key: string, value: unknown) => {
    onChange({ ...tool, [key]: value })
  }

  const setConfig = (key: string, value: unknown) => {
    onChange({ ...tool, config: { ...tool.config, [key]: value } })
  }

  const cfg = (key: string, fallback = '') => (tool.config[key] as string) ?? fallback

  return (
    <div className="border border-white/[0.08] rounded-xl p-4 mb-3 bg-white/[0.02]">
      <div className="flex items-center justify-between mb-3">
        <span className="text-white/60 text-sm font-semibold">
          {tool.name || t.agentModal.toolName}
        </span>
        <button
          onClick={onDelete}
          className="text-danger/60 hover:text-danger text-xs transition-colors"
        >
          {t.admin.delete}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-3 mb-3">
        <div>
          <label className={lbl}>{t.agentModal.toolName}</label>
          <input
            className={inp}
            value={tool.name}
            onChange={(e) => set('name', e.target.value)}
            placeholder="e.g. coze_workflow"
          />
        </div>
        <div>
          <label className={lbl}>{t.agentModal.toolType}</label>
          <select
            className={inp}
            value={tool.type}
            onChange={(e) => set('type', e.target.value)}
          >
            <option value="coze">Coze</option>
            <option value="n8n">n8n</option>
          </select>
        </div>
      </div>

      <div className="mb-3">
        <label className={lbl}>{t.agentModal.toolDesc}</label>
        <input
          className={inp}
          value={tool.description}
          onChange={(e) => set('description', e.target.value)}
          placeholder="Describe what this tool does..."
        />
      </div>

      <div className="mb-3">
        <label className={lbl}>{t.agentModal.toolInputSchema}</label>
        <textarea
          className={`${inp} resize-none font-mono text-xs`}
          rows={3}
          value={tool.input_schema}
          onChange={(e) => set('input_schema', e.target.value)}
        />
      </div>

      <div className="border-t border-white/[0.06] pt-3">
        {tool.type === 'coze' && (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={lbl}>{t.agentModal.cozeWorkflowId}</label>
                <input
                  className={inp}
                  value={cfg('workflow_id')}
                  onChange={(e) => setConfig('workflow_id', e.target.value)}
                />
              </div>
              <div>
                <label className={lbl}>{t.agentModal.cozeApiKey}</label>
                <input
                  type="password"
                  className={inp}
                  value={cfg('api_key')}
                  onChange={(e) => setConfig('api_key', e.target.value)}
                  placeholder="pat_..."
                />
              </div>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className={lbl}>{t.agentModal.cozeRegion}</label>
                <select
                  className={inp}
                  value={cfg('region', 'cn')}
                  onChange={(e) => setConfig('region', e.target.value)}
                >
                  <option value="cn">{t.agentModal.cozeRegionCn}</option>
                  <option value="global">{t.agentModal.cozeRegionGlobal}</option>
                </select>
              </div>
              <div>
                <label className={lbl}>{t.agentModal.cozeInputField}</label>
                <input
                  className={inp}
                  value={cfg('input_field')}
                  onChange={(e) => setConfig('input_field', e.target.value)}
                  placeholder="input"
                />
              </div>
              <div>
                <label className={lbl}>{t.agentModal.cozeOutputField}</label>
                <input
                  className={inp}
                  value={cfg('output_field')}
                  onChange={(e) => setConfig('output_field', e.target.value)}
                  placeholder="output"
                />
              </div>
            </div>
          </div>
        )}

        {tool.type === 'n8n' && (
          <div className="space-y-3">
            <div>
              <label className={lbl}>{t.agentModal.n8nWebhookUrl}</label>
              <input
                className={`${inp} font-mono text-xs`}
                value={cfg('webhook_url')}
                onChange={(e) => setConfig('webhook_url', e.target.value)}
                placeholder="https://n8n.example.com/webhook/..."
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={lbl}>{t.agentModal.n8nAuthType}</label>
                <select
                  className={inp}
                  value={cfg('auth_type', 'none')}
                  onChange={(e) => setConfig('auth_type', e.target.value)}
                >
                  <option value="none">{t.agentModal.n8nAuthNone}</option>
                  <option value="bearer">{t.agentModal.n8nAuthBearer}</option>
                  <option value="header">{t.agentModal.n8nAuthHeader}</option>
                </select>
              </div>
              {cfg('auth_type') && cfg('auth_type') !== 'none' && (
                <div>
                  <label className={lbl}>{t.agentModal.n8nAuthToken}</label>
                  <input
                    type="password"
                    className={inp}
                    value={cfg('auth_token')}
                    onChange={(e) => setConfig('auth_token', e.target.value)}
                  />
                </div>
              )}
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={lbl}>{t.agentModal.n8nTimeout}</label>
                <input
                  type="number"
                  className={inp}
                  value={cfg('timeout', '30')}
                  onChange={(e) => setConfig('timeout', Number(e.target.value))}
                />
              </div>
              <div>
                <label className={lbl}>{t.agentModal.n8nPayloadTmpl}</label>
                <input
                  className={`${inp} font-mono text-xs`}
                  value={cfg('payload_tmpl')}
                  onChange={(e) => setConfig('payload_tmpl', e.target.value)}
                  placeholder='{"input": "{{.input}}"}'
                />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
