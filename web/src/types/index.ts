export interface User {
  id: number
  email: string
  name: string
  role: string
  status: string
  created_at: string
}

export interface Agent {
  id: number
  name: string
  short_desc: string
  full_desc: string
  description: string
  icon: string
  color: string
  category: string
  tags: string[]
  status: string
  prompt: string
  model_name: string
  base_url: string
  api_key: string
  temperature: number
  max_tokens: number
  cost: number
  calls: number
  tools: AgentTool[]
  created_at: string
  updated_at: string
}

export interface AgentTool {
  id: number
  agent_id: number
  name: string
  type: string
  description: string
  config: string
  input_schema: string
}

export interface Conversation {
  id: number
  user_id: number
  agent_id: number
  title: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: number
  conversation_id: number
  role: string
  content: string
  tool_calls: string
  tool_call_id: string
  created_at: string
}

export interface KnowledgeBase {
  id: number
  name: string
  description: string
  status: string
  embedding_base_url: string
  embedding_api_key: string
  embedding_model: string
  dimension: number
  chunk_size: number
  chunk_overlap: number
  created_at: string
  updated_at: string
}

export interface Document {
  id: number
  knowledge_base_id: number
  name: string
  content: string
  status: string
  chunk_count: number
  created_at: string
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  meta?: {
    total: number
    page: number
    limit: number
  }
}

export interface LoginResponse {
  token: string
  user: User
}

export interface SendMessageResponse {
  conversation_id: number
  reply: string
}
