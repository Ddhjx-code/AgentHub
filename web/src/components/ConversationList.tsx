import { useLocale } from '../contexts/LocaleContext'
import type { Conversation } from '../types'

interface Props {
  conversations: Conversation[]
  activeId: number | null
  onSelect: (conv: Conversation) => void
  onDelete: (id: number) => void
  onNew: () => void
}

export default function ConversationList({ conversations, activeId, onSelect, onDelete, onNew }: Props) {
  const { t } = useLocale()

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-white/[0.06]">
        <button
          onClick={onNew}
          className="w-full py-2.5 bg-primary/15 border border-primary/25 text-primary text-sm font-semibold rounded-xl hover:bg-primary/25 transition-colors"
        >
          {t.chat.newChat}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {conversations.length === 0 ? (
          <p className="text-white/25 text-xs text-center py-8">{t.chat.noConversations}</p>
        ) : (
          conversations.map((conv) => (
            <div
              key={conv.id}
              onClick={() => onSelect(conv)}
              className={`group flex items-center gap-2 px-3 py-2.5 rounded-xl cursor-pointer transition-all ${
                activeId === conv.id
                  ? 'bg-primary/15 border border-primary/20'
                  : 'hover:bg-white/[0.04] border border-transparent'
              }`}
            >
              <div className="flex-1 min-w-0">
                <p className={`text-sm truncate ${activeId === conv.id ? 'text-primary' : 'text-white/60'}`}>
                  {conv.title || 'Untitled'}
                </p>
                <p className="text-white/20 text-[10px]">
                  {new Date(conv.created_at).toLocaleDateString()}
                </p>
              </div>
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onDelete(conv.id)
                }}
                className="opacity-0 group-hover:opacity-100 text-white/25 hover:text-danger text-xs transition-all"
              >
                x
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
