import { useState, useMemo } from 'react'
import ReactMarkdown from 'react-markdown'
import { useLocale } from '../contexts/LocaleContext'
import type { Message } from '../types'

interface ToolCall {
  id: string
  type: string
  function: {
    name: string
    arguments: string
  }
}

function formatArguments(args: string): string {
  try {
    return JSON.stringify(JSON.parse(args), null, 2)
  } catch {
    return args
  }
}

interface Props {
  message: Message
}

export default function ChatMessage({ message }: Props) {
  const { t } = useLocale()
  const isUser = message.role === 'user'
  const isTool = message.role === 'tool'
  const [expanded, setExpanded] = useState(false)

  const toolCalls: ToolCall[] = useMemo(() => {
    if (message.role !== 'assistant' || !message.tool_calls) return []
    try {
      return JSON.parse(message.tool_calls)
    } catch {
      return []
    }
  }, [message.role, message.tool_calls])

  if (isTool) {
    return (
      <div className="flex justify-start mb-4">
        <div className="max-w-[80%] px-4 py-2 bg-secondary/10 border border-secondary/20 rounded-xl text-xs text-white/50 font-mono">
          <span className="text-secondary text-[10px] uppercase tracking-wider">{t.chat.toolResult}</span>
          <p className="mt-1 whitespace-pre-wrap">{message.content}</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
      <div
        className={`max-w-[80%] px-4 py-3 rounded-2xl ${
          isUser
            ? 'bg-primary/20 border border-primary/30 text-white'
            : 'bg-white/[0.04] border border-white/[0.08] text-white/80'
        }`}
      >
        {isUser ? (
          <p className="text-sm whitespace-pre-wrap">{message.content}</p>
        ) : (
          <>
            {message.content && (
              <div className="prose prose-invert prose-sm max-w-none [&_p]:text-white/80 [&_p]:text-sm [&_code]:bg-white/10 [&_code]:px-1 [&_code]:rounded [&_pre]:bg-white/5 [&_pre]:rounded-xl [&_pre]:p-3 [&_a]:text-primary [&_h1]:text-white [&_h2]:text-white [&_h3]:text-white [&_li]:text-white/80">
                <ReactMarkdown>{message.content}</ReactMarkdown>
              </div>
            )}
            {toolCalls.length > 0 && (
              <div className={`${message.content ? 'mt-2 border-t border-white/[0.06] pt-2' : ''}`}>
                <button
                  onClick={() => setExpanded(!expanded)}
                  className="text-xs text-secondary/70 hover:text-secondary flex items-center gap-1 transition-colors"
                >
                  <span>{expanded ? '▼' : '▶'}</span>
                  <span>{t.chat.toolCall} ({toolCalls.length})</span>
                  <span className="text-white/20 ml-1">
                    {expanded ? t.chat.toolCallCollapse : t.chat.toolCallExpand}
                  </span>
                </button>
                {expanded && (
                  <div className="mt-2 space-y-2">
                    {toolCalls.map((tc) => (
                      <div key={tc.id} className="bg-white/[0.03] border border-white/[0.06] rounded-lg p-2">
                        <p className="text-xs font-semibold text-secondary/80">{tc.function.name}</p>
                        <pre className="text-[10px] text-white/40 mt-1 whitespace-pre-wrap overflow-x-auto font-mono">
                          {formatArguments(tc.function.arguments)}
                        </pre>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </>
        )}
        <p className={`text-[10px] mt-2 ${isUser ? 'text-primary/50' : 'text-white/20'}`}>
          {new Date(message.created_at).toLocaleTimeString()}
        </p>
      </div>
    </div>
  )
}
