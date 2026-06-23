import client from './client'
import type { Agent, ApiResponse } from '../types'

interface ListAgentsParams {
  page?: number
  limit?: number
  category?: string
  tag?: string
}

interface ListAgentsResponse {
  agents: Agent[]
  meta: { total: number; page: number; limit: number }
}

export async function listAgents(params: ListAgentsParams = {}): Promise<ListAgentsResponse> {
  const resp = await client.get<ApiResponse<Agent[]>>('/agents', { params })
  return {
    agents: resp.data.data || [],
    meta: resp.data.meta || { total: 0, page: 1, limit: 20 },
  }
}

export async function getAgent(id: number): Promise<Agent> {
  const resp = await client.get<ApiResponse<Agent>>(`/agents/${id}`)
  return resp.data.data
}

export async function createAgent(data: Partial<Agent>): Promise<Agent> {
  const resp = await client.post<ApiResponse<Agent>>('/admin/agents', data)
  return resp.data.data
}

export async function updateAgent(id: number, data: Partial<Agent>): Promise<Agent> {
  const resp = await client.put<ApiResponse<Agent>>(`/admin/agents/${id}`, data)
  return resp.data.data
}

export async function deleteAgent(id: number): Promise<void> {
  await client.delete(`/admin/agents/${id}`)
}

export async function toggleAgent(id: number): Promise<void> {
  await client.put(`/admin/agents/${id}/toggle`)
}

export async function listAllAgents(params: ListAgentsParams = {}): Promise<ListAgentsResponse> {
  const resp = await client.get<ApiResponse<Agent[]>>('/admin/agents', { params })
  return {
    agents: resp.data.data || [],
    meta: resp.data.meta || { total: 0, page: 1, limit: 20 },
  }
}
