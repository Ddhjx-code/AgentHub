import ReactMarkdown from 'react-markdown'
import type { Message } from '../types'

interface Props {
  message: Message
}

export default function ChatMessage({ message }: Props) {
  const isUser = message.role === 'user'
  const isTool = message.role === 'tool'

  if (isTool) {
    return (
      <div className="flex justify-start mb-4">
        <div className="max-w-[80%] px-4 py-2 bg-secondary/10 border border-secondary/20 rounded-xl text-xs text-white/50 font-mono">
          <span className="text-secondary text-[10px] uppercase tracking-wider">Tool Result</span>
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
          <div className="prose prose-invert prose-sm max-w-none [&_p]:text-white/80 [&_p]:text-sm [&_code]:bg-white/10 [&_code]:px-1 [&_code]:rounded [&_pre]:bg-white/5 [&_pre]:rounded-xl [&_pre]:p-3 [&_a]:text-primary [&_h1]:text-white [&_h2]:text-white [&_h3]:text-white [&_li]:text-white/80">
            <ReactMarkdown>{message.content}</ReactMarkdown>
          </div>
        )}
        <p className={`text-[10px] mt-2 ${isUser ? 'text-primary/50' : 'text-white/20'}`}>
          {new Date(message.created_at).toLocaleTimeString()}
        </p>
      </div>
    </div>
  )
}
