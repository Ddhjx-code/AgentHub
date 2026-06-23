import client from './client'
import type { ApiResponse, Conversation, Message, SendMessageResponse } from '../types'

export async function sendMessage(
  agentId: number,
  content: string,
  conversationId?: number,
): Promise<SendMessageResponse> {
  const body: Record<string, unknown> = { agent_id: agentId, content }
  if (conversationId) {
    body.conversation_id = conversationId
  }
  const resp = await client.post<ApiResponse<SendMessageResponse>>(`/agents/${agentId}/chat`, body)
  return resp.data.data
}

export async function listConversations(): Promise<Conversation[]> {
  const resp = await client.get<ApiResponse<Conversation[]>>('/conversations')
  return resp.data.data || []
}

export async function getMessages(conversationId: number): Promise<Message[]> {
  const resp = await client.get<ApiResponse<Message[]>>(`/conversations/${conversationId}/messages`)
  return resp.data.data || []
}

export async function deleteConversation(conversationId: number): Promise<void> {
  await client.delete(`/conversations/${conversationId}`)
}
