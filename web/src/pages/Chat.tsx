import { useState, useEffect, useRef } from 'react'
import { useParams, useSearchParams, Link } from 'react-router-dom'
import { getAgent } from '../api/agent'
import { sendMessage, listConversations, getMessages, deleteConversation } from '../api/chat'
import type { Agent, Conversation, Message } from '../types'
import ConversationList from '../components/ConversationList'
import ChatMessage from '../components/ChatMessage'
import { useAuth } from '../contexts/AuthContext'

export default function Chat() {
  const { agentId } = useParams<{ agentId: string }>()
  const [searchParams] = useSearchParams()
  const { flash } = useAuth()

  const [agent, setAgent] = useState<Agent | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeConvId, setActiveConvId] = useState<number | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!agentId) return
    getAgent(Number(agentId)).then(setAgent).catch(() => {})
    loadConversations()
    const convParam = searchParams.get('conversation')
    if (convParam) setActiveConvId(Number(convParam))
  }, [agentId, searchParams])

  useEffect(() => {
    if (activeConvId) {
      getMessages(activeConvId).then(setMessages).catch(() => {})
    } else {
      setMessages([])
    }
  }, [activeConvId])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const loadConversations = () => {
    listConversations()
      .then((convs) => {
        const filtered = convs.filter((c) => c.agent_id === Number(agentId))
        setConversations(filtered)
      })
      .catch(() => {})
  }

  const handleSend = async () => {
    if (!input.trim() || !agentId || sending) return
    const content = input.trim()
    setInput('')
    setSending(true)

    const tempUserMsg: Message = {
      id: Date.now(),
      conversation_id: activeConvId || 0,
      role: 'user',
      content,
      tool_calls: '',
      tool_call_id: '',
      created_at: new Date().toISOString(),
    }
    setMessages((prev) => [...prev, tempUserMsg])

    try {
      const resp = await sendMessage(Number(agentId), content, activeConvId || undefined)
      if (!activeConvId) {
        setActiveConvId(resp.conversation_id)
        loadConversations()
      }

      const tempAiMsg: Message = {
        id: Date.now() + 1,
        conversation_id: resp.conversation_id,
        role: 'assistant',
        content: resp.reply,
        tool_calls: '',
        tool_call_id: '',
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, tempAiMsg])
    } catch (err) {
      flash(err instanceof Error ? err.message : 'Failed to send message', 'error')
    } finally {
      setSending(false)
    }
  }

  const handleDeleteConv = async (id: number) => {
    try {
      await deleteConversation(id)
      if (activeConvId === id) {
        setActiveConvId(null)
        setMessages([])
      }
      loadConversations()
    } catch {
      flash('Failed to delete conversation', 'error')
    }
  }

  const handleNewChat = () => {
    setActiveConvId(null)
    setMessages([])
  }

  return (
    <div className="h-[calc(100vh-4rem)] flex">
      {/* Sidebar */}
      <div className="w-64 shrink-0 bg-white/[0.02] border-r border-white/[0.06] hidden md:block">
        <div className="p-4 border-b border-white/[0.06]">
          <Link to={`/agents/${agentId}`} className="flex items-center gap-2 text-sm text-white/45 hover:text-white/75 transition-colors">
            <span className="text-lg">{agent?.icon || '🤖'}</span>
            <span className="truncate font-medium">{agent?.name || 'Agent'}</span>
          </Link>
        </div>
        <ConversationList
          conversations={conversations}
          activeId={activeConvId}
          onSelect={(conv) => setActiveConvId(conv.id)}
          onDelete={handleDeleteConv}
          onNew={handleNewChat}
        />
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Messages */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {messages.length === 0 ? (
            <div className="h-full flex items-center justify-center">
              <div className="text-center">
                <div className="text-5xl mb-4">{agent?.icon || '🤖'}</div>
                <h2 className="text-xl font-bold text-white mb-2">{agent?.name || 'Agent'}</h2>
                <p className="text-white/35 text-sm max-w-md">{agent?.description || 'Start a conversation'}</p>
              </div>
            </div>
          ) : (
            <>
              {messages.map((msg) => (
                <ChatMessage key={msg.id} message={msg} />
              ))}
              {sending && (
                <div className="flex justify-start mb-4">
                  <div className="px-4 py-3 bg-white/[0.04] border border-white/[0.08] rounded-2xl">
                    <div className="flex items-center gap-2 text-white/40 text-sm">
                      <span className="animate-pulse">&#9679;</span>
                      <span className="animate-pulse" style={{ animationDelay: '0.2s' }}>&#9679;</span>
                      <span className="animate-pulse" style={{ animationDelay: '0.4s' }}>&#9679;</span>
                    </div>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </>
          )}
        </div>

        {/* Input */}
        <div className="border-t border-white/[0.06] p-4">
          <div className="flex gap-3 max-w-4xl mx-auto">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() } }}
              placeholder="Type your message..."
              disabled={sending}
              className="flex-1 bg-white/[0.04] border border-white/[0.08] rounded-xl px-4 py-3 text-white placeholder-white/25 text-sm focus:outline-none focus:border-primary/40 transition-colors disabled:opacity-50"
            />
            <button
              onClick={handleSend}
              disabled={sending || !input.trim()}
              className="px-6 py-3 bg-primary text-base font-bold rounded-xl hover:bg-[#00bfe8] transition-colors shadow-[0_0_15px_rgba(0,212,255,0.2)] disabled:opacity-50 disabled:shadow-none shrink-0"
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
