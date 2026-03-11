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

// Phase 3: MessagePart union types for tool-aware chat display

export interface TextPart {
  type: 'text'
  content: string
}

export interface ToolCallPart {
  type: 'tool_call'
  id: string
  name: string
  args: string  // raw JSON string from backend
}

export interface ToolResultPart {
  type: 'tool_result'
  id: string     // matches ToolCallPart.id
  content: string
  error?: string
}

export interface ApprovalRequestPart {
  type: 'approval_request'
  approvalId: string
  tool: string
  command: string
}

export type MessagePart = TextPart | ToolCallPart | ToolResultPart | ApprovalRequestPart

// ChatMessage is used during live streaming (replaces flat Message for accumulation).
// Historical messages loaded from backend use Message (flat content string) and are
// converted to ChatMessage on load as a single TextPart.
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system'
  parts: MessagePart[]
}

// submitApproval posts the user's approval decision to the backend.
export async function submitApproval(approvalId: string, approved: boolean): Promise<void> {
  const res = await fetch(`${API_BASE}/api/chat/approve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ approvalId, approved }),
  })
  // 204 = success; any other status = silently ignore (stream has already timed out)
  if (!res.ok && res.status !== 404) {
    throw new Error(`approve failed: ${res.status}`)
  }
}
