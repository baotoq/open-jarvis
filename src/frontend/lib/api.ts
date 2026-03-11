export const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8888'

export interface Conversation {
  id: string
  title: string
  createdAt: number   // Unix timestamp seconds
  updatedAt: number   // Unix timestamp seconds
}

export interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export async function listConversations(): Promise<Conversation[]> {
  const res = await fetch(`${API_BASE}/api/conversations`)
  if (!res.ok) throw new Error(`list failed: ${res.status}`)
  return res.json()
}

export async function getConversation(id: string): Promise<Conversation | null> {
  const res = await fetch(`${API_BASE}/api/conversations/${id}`)
  if (res.status === 404) return null
  if (!res.ok) throw new Error(`get failed: ${res.status}`)
  return res.json()
}

export async function getConversationMessages(id: string): Promise<Message[]> {
  const res = await fetch(`${API_BASE}/api/conversations/${id}/messages`)
  if (!res.ok) throw new Error(`messages failed: ${res.status}`)
  return res.json()
}

export async function deleteConversation(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/conversations/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`delete failed: ${res.status}`)
}
