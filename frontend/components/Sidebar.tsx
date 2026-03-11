'use client'
import { useEffect, useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { listConversations, deleteConversation, type Conversation } from '@/lib/api'

interface ConvEntryProps {
  conv: Conversation
  active: boolean
  onClick: () => void
  onDelete: (id: string) => void
}

function ConvEntry({ conv, active, onClick, onDelete }: ConvEntryProps) {
  const relative = formatDistanceToNow(new Date(conv.updatedAt * 1000), { addSuffix: true })
  return (
    <div
      onClick={onClick}
      className={cn(
        'flex flex-col px-3 py-2 rounded-md cursor-pointer select-none group relative',
        active ? 'bg-accent' : 'hover:bg-muted'
      )}
    >
      <span className="text-sm font-medium truncate pr-8">{conv.title || 'New conversation'}</span>
      <span className="text-xs text-muted-foreground">{relative}</span>
      <button
        className="absolute right-2 top-2 opacity-0 group-hover:opacity-100 text-xs text-destructive"
        onClick={(e) => { e.stopPropagation(); onDelete(conv.id) }}
      >
        Delete
      </button>
    </div>
  )
}

interface SidebarProps {
  activeId: string | null
  onSelect: (id: string) => void
  onNew: () => void
}

export function Sidebar({ activeId, onSelect, onNew }: SidebarProps) {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState(false)

  const fetchConversations = () => {
    setIsLoading(true)
    setError(false)
    listConversations()
      .then(convs => setConversations(convs))
      .catch(() => setError(true))
      .finally(() => setIsLoading(false))
  }

  useEffect(() => {
    fetchConversations()
  }, [activeId])

  const handleDelete = async (id: string) => {
    try {
      await deleteConversation(id)
      if (id === activeId) {
        onNew()
      }
      fetchConversations()
    } catch {
      // ignore delete errors
    }
  }

  return (
    <div className="w-64 flex-shrink-0 border-r bg-background flex flex-col h-screen">
      <div className="flex items-center justify-between p-3 border-b">
        <span className="font-semibold text-lg">Jarvis</span>
        <Button size="sm" variant="outline" onClick={onNew}>
          New chat
        </Button>
      </div>
      <div className="overflow-y-auto flex-1 p-2 flex flex-col gap-1">
        {isLoading && (
          <p className="text-sm text-muted-foreground px-3 py-2">Loading...</p>
        )}
        {!isLoading && error && (
          <p className="text-sm text-muted-foreground px-3 py-2">Could not load conversations</p>
        )}
        {!isLoading && !error && conversations.map(conv => (
          <ConvEntry
            key={conv.id}
            conv={conv}
            active={conv.id === activeId}
            onClick={() => onSelect(conv.id)}
            onDelete={handleDelete}
          />
        ))}
      </div>
    </div>
  )
}
