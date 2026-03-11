'use client'
import { useEffect, useRef, useState } from 'react'
import {
  API_BASE,
  getConversationMessages,
  submitApproval,
  type ChatMessage,
  type ToolCallPart,
  type ToolResultPart,
  type ApprovalRequestPart,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { ToolCallBlock } from '@/components/ToolCallBlock'
import { ApprovalDialog } from '@/components/ApprovalDialog'

interface ChatAreaProps {
  sessionId: string | null
  onSessionCreated: (id: string) => void
}

export function ChatArea({ sessionId, onSessionCreated }: ChatAreaProps) {
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([])
  const [pendingApproval, setPendingApproval] = useState<ApprovalRequestPart | null>(null)
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!sessionId) {
      setChatMessages([])
      return
    }
    setIsLoadingHistory(true)
    getConversationMessages(sessionId)
      .then(msgs => setChatMessages(msgs.map(m => ({
        role: m.role,
        parts: [{ type: 'text' as const, content: m.content }],
      }))))
      .catch(() => setChatMessages([]))
      .finally(() => setIsLoadingHistory(false))
  }, [sessionId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatMessages])

  const handleSend = async () => {
    const message = input.trim()
    if (!message || isStreaming || isLoadingHistory) return

    setChatMessages(prev => [...prev, { role: 'user', parts: [{ type: 'text', content: message }] }])
    setInput('')
    setIsStreaming(true)

    try {
      const res = await fetch(`${API_BASE}/api/chat/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId: sessionId ?? '', message }),
      })

      if (!res.ok || !res.body) {
        setIsStreaming(false)
        return
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let assistantContent = ''

      // Add empty assistant message to append tokens into
      setChatMessages(prev => [...prev, { role: 'assistant', parts: [{ type: 'text', content: '' }] }])

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice('data: '.length)
          if (!data) continue

          try {
            const parsed = JSON.parse(data)
            if (parsed.done === true && parsed.sessionId) {
              onSessionCreated(parsed.sessionId)
              break
            }
            if (parsed.type === 'tool_call') {
              setChatMessages(prev => {
                const updated = [...prev]
                const last = { ...updated[updated.length - 1] }
                last.parts = [...last.parts, { type: 'tool_call', id: parsed.id, name: parsed.name, args: parsed.args } as ToolCallPart]
                updated[updated.length - 1] = last
                return updated
              })
            } else if (parsed.type === 'tool_result') {
              setChatMessages(prev => {
                const updated = [...prev]
                const last = { ...updated[updated.length - 1] }
                last.parts = [...last.parts, { type: 'tool_result', id: parsed.id, content: parsed.content, error: parsed.error } as ToolResultPart]
                updated[updated.length - 1] = last
                return updated
              })
            } else if (parsed.type === 'approval_request') {
              const part: ApprovalRequestPart = { type: 'approval_request', approvalId: parsed.approvalId, tool: parsed.tool, command: parsed.command }
              setChatMessages(prev => {
                const updated = [...prev]
                const last = { ...updated[updated.length - 1] }
                last.parts = [...last.parts, part]
                updated[updated.length - 1] = last
                return updated
              })
              setPendingApproval(part)
            }
          } catch {
            // bare text token
            assistantContent += data
            setChatMessages(prev => {
              const updated = [...prev]
              const last = { ...updated[updated.length - 1] }
              const parts = [...last.parts]
              const lastPart = parts[parts.length - 1]
              if (lastPart?.type === 'text') {
                parts[parts.length - 1] = { type: 'text', content: assistantContent }
              } else {
                parts.push({ type: 'text', content: assistantContent })
              }
              last.parts = parts
              updated[updated.length - 1] = last
              return updated
            })
          }
        }
      }
    } catch {
      // Stream error — leave messages as-is
    } finally {
      setIsStreaming(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full">
      <ApprovalDialog
        open={pendingApproval !== null}
        command={pendingApproval?.command ?? ''}
        onApprove={async () => {
          if (!pendingApproval) return
          await submitApproval(pendingApproval.approvalId, true)
          setPendingApproval(null)
        }}
        onDeny={async () => {
          if (!pendingApproval) return
          await submitApproval(pendingApproval.approvalId, false)
          setPendingApproval(null)
        }}
      />
      {isLoadingHistory ? (
        <div className="flex-1 flex items-center justify-center">
          <p className="text-muted-foreground text-sm">Loading conversation...</p>
        </div>
      ) : (
        <div className="flex flex-col gap-2 flex-1 overflow-y-auto p-4">
          {chatMessages.length === 0 && (
            <div className="flex-1 flex items-center justify-center">
              <p className="text-muted-foreground text-sm">Start a conversation...</p>
            </div>
          )}
          {chatMessages.map((msg, i) => (
            <div
              key={i}
              className={`flex flex-col gap-1 max-w-2xl ${msg.role === 'user' ? 'self-end items-end' : 'self-start items-start w-full'}`}
            >
              <span className="text-xs text-muted-foreground capitalize">{msg.role}</span>
              {msg.parts.map((part, j) => {
                if (part.type === 'text') {
                  return (
                    <div
                      key={j}
                      className={`rounded-lg px-4 py-2 text-sm whitespace-pre-wrap ${
                        msg.role === 'user'
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-muted text-foreground'
                      }`}
                    >
                      {part.content}
                    </div>
                  )
                }
                if (part.type === 'tool_call') {
                  // Find matching tool_result from same message
                  const resultPart = msg.parts.find(
                    p => p.type === 'tool_result' && p.id === part.id
                  ) as ToolResultPart | undefined
                  return <ToolCallBlock key={j} call={part} result={resultPart} />
                }
                if (part.type === 'tool_result') {
                  // Already rendered via ToolCallBlock above; skip standalone
                  return null
                }
                if (part.type === 'approval_request') {
                  return (
                    <div key={j} className="text-xs text-muted-foreground italic">
                      Waiting for approval: <span className="font-mono">{part.command}</span>
                    </div>
                  )
                }
                return null
              })}
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      )}
      <div className="border-t p-4 flex gap-2">
        <textarea
          className="flex-1 resize-none rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring min-h-[60px]"
          placeholder="Type a message... (Enter to send, Shift+Enter for newline)"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={isStreaming || isLoadingHistory}
        />
        <Button
          onClick={handleSend}
          disabled={isStreaming || isLoadingHistory || !input.trim()}
        >
          {isStreaming ? 'Sending...' : 'Send'}
        </Button>
      </div>
    </div>
  )
}
