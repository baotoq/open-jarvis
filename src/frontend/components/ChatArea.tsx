'use client'
import { useEffect, useRef, useState } from 'react'
import { API_BASE, getConversationMessages, type Message } from '@/lib/api'
import { Button } from '@/components/ui/button'

interface ChatAreaProps {
  sessionId: string | null
  onSessionCreated: (id: string) => void
}

export function ChatArea({ sessionId, onSessionCreated }: ChatAreaProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!sessionId) {
      setMessages([])
      return
    }
    setIsLoadingHistory(true)
    getConversationMessages(sessionId)
      .then(msgs => setMessages(msgs))
      .catch(() => setMessages([]))
      .finally(() => setIsLoadingHistory(false))
  }, [sessionId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    const message = input.trim()
    if (!message || isStreaming || isLoadingHistory) return

    setMessages(prev => [...prev, { role: 'user', content: message }])
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
      setMessages(prev => [...prev, { role: 'assistant', content: '' }])

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice('data: '.length)
          if (!data) continue

          // Try to parse as done event
          try {
            const parsed = JSON.parse(data)
            if (parsed.done === true && parsed.sessionId) {
              onSessionCreated(parsed.sessionId)
              break
            }
          } catch {
            // Not JSON — it's a text token
            assistantContent += data
            setMessages(prev => {
              const updated = [...prev]
              updated[updated.length - 1] = { role: 'assistant', content: assistantContent }
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
      {isLoadingHistory ? (
        <div className="flex-1 flex items-center justify-center">
          <p className="text-muted-foreground text-sm">Loading conversation...</p>
        </div>
      ) : (
        <div className="flex flex-col gap-2 flex-1 overflow-y-auto p-4">
          {messages.length === 0 && (
            <div className="flex-1 flex items-center justify-center">
              <p className="text-muted-foreground text-sm">Start a conversation...</p>
            </div>
          )}
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`flex flex-col gap-1 max-w-2xl ${msg.role === 'user' ? 'self-end items-end' : 'self-start items-start'}`}
            >
              <span className="text-xs text-muted-foreground capitalize">{msg.role}</span>
              <div
                className={`rounded-lg px-4 py-2 text-sm whitespace-pre-wrap ${
                  msg.role === 'user'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-foreground'
                }`}
              >
                {msg.content}
              </div>
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
