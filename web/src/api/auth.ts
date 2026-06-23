import client from './client'
import type { ApiResponse, LoginResponse, User } from '../types'

export async function login(email: string, password: string): Promise<LoginResponse> {
  const resp = await client.post<ApiResponse<LoginResponse>>('/auth/login', { email, password })
  return resp.data.data
}

export async function register(email: string, name: string, password: string): Promise<User> {
  const resp = await client.post<ApiResponse<User>>('/auth/register', { email, name, password })
  return resp.data.data
}

export async function getProfile(): Promise<User> {
  const resp = await client.get<ApiResponse<User>>('/user/profile')
  return resp.data.data
}
