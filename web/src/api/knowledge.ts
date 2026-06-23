import client from './client'
import type { ApiResponse, KnowledgeBase, Document } from '../types'

export async function listKnowledgeBases(): Promise<KnowledgeBase[]> {
  const resp = await client.get<ApiResponse<KnowledgeBase[]>>('/admin/knowledge-bases')
  return resp.data.data || []
}

export async function getKnowledgeBase(id: number): Promise<KnowledgeBase> {
  const resp = await client.get<ApiResponse<KnowledgeBase>>(`/admin/knowledge-bases/${id}`)
  return resp.data.data
}

export async function createKnowledgeBase(data: Partial<KnowledgeBase>): Promise<KnowledgeBase> {
  const resp = await client.post<ApiResponse<KnowledgeBase>>('/admin/knowledge-bases', data)
  return resp.data.data
}

export async function updateKnowledgeBase(id: number, data: Partial<KnowledgeBase>): Promise<KnowledgeBase> {
  const resp = await client.put<ApiResponse<KnowledgeBase>>(`/admin/knowledge-bases/${id}`, data)
  return resp.data.data
}

export async function deleteKnowledgeBase(id: number): Promise<void> {
  await client.delete(`/admin/knowledge-bases/${id}`)
}

export async function listDocuments(kbId: number): Promise<Document[]> {
  const resp = await client.get<ApiResponse<Document[]>>(`/admin/knowledge-bases/${kbId}/documents`)
  return resp.data.data || []
}

export async function uploadDocument(kbId: number, name: string, content: string): Promise<Document> {
  const resp = await client.post<ApiResponse<Document>>(`/admin/knowledge-bases/${kbId}/documents`, { name, content })
  return resp.data.data
}

export async function deleteDocument(kbId: number, docId: number): Promise<void> {
  await client.delete(`/admin/knowledge-bases/${kbId}/documents/${docId}`)
}

export async function bindAgentKB(agentId: number, kbId: number): Promise<void> {
  await client.post(`/admin/agents/${agentId}/knowledge-bases`, { knowledge_base_id: kbId })
}

export async function unbindAgentKB(agentId: number, kbId: number): Promise<void> {
  await client.delete(`/admin/agents/${agentId}/knowledge-bases/${kbId}`)
}

export async function listAgentKBs(agentId: number): Promise<KnowledgeBase[]> {
  const resp = await client.get<ApiResponse<KnowledgeBase[]>>(`/admin/agents/${agentId}/knowledge-bases`)
  return resp.data.data || []
}
