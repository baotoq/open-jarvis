'use client'
import { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Settings2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  listConversations, deleteConversation, searchConversations,
  type Conversation, type SearchResult
} from '@/lib/api'

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

function SearchResultEntry({ result, onClick }: { result: SearchResult; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      className="flex flex-col px-3 py-2 rounded-md cursor-pointer select-none hover:bg-muted"
    >
      <span className="text-sm font-medium truncate">{result.title || 'Untitled'}</span>
      <span
        className="text-xs text-muted-foreground line-clamp-2"
        dangerouslySetInnerHTML={{ __html: result.snippet }}
      />
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
  const [query, setQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(null)
  const [isSearching, setIsSearching] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

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

  const handleSearch = (q: string) => {
    setQuery(q)
    clearTimeout(timerRef.current)
    if (!q.trim()) {
      setSearchResults(null)
      return
    }
    timerRef.current = setTimeout(async () => {
      setIsSearching(true)
      try {
        const results = await searchConversations(q)
        setSearchResults(results)
      } catch {
        setSearchResults([])
      } finally {
        setIsSearching(false)
      }
    }, 300)
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteConversation(id)
      if (id === activeId) onNew()
      fetchConversations()
    } catch {
      // ignore delete errors
    }
  }

  const showSearch = query.trim().length > 0

  return (
    <div className="w-64 flex-shrink-0 border-r bg-background flex flex-col h-screen">
      <div className="flex items-center justify-between p-3 border-b">
        <span className="font-semibold text-lg">Jarvis</span>
        <div className="flex items-center gap-1">
          <Link href="/settings" className="p-1 rounded hover:bg-muted" title="Settings">
            <Settings2 className="w-4 h-4 text-muted-foreground" />
          </Link>
          <Button size="sm" variant="outline" onClick={onNew}>New chat</Button>
        </div>
      </div>
      <div className="px-2 pt-2 pb-1">
        <input
          type="text"
          value={query}
          onChange={e => handleSearch(e.target.value)}
          placeholder="Search conversations..."
          className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>
      <div className="overflow-y-auto flex-1 p-2 flex flex-col gap-1">
        {showSearch ? (
          <>
            {isSearching && <p className="text-sm text-muted-foreground px-3 py-2">Searching...</p>}
            {!isSearching && searchResults !== null && searchResults.length === 0 && (
              <p className="text-sm text-muted-foreground px-3 py-2">No results</p>
            )}
            {!isSearching && searchResults && searchResults.map(r => (
              <SearchResultEntry key={r.id} result={r} onClick={() => onSelect(r.id)} />
            ))}
          </>
        ) : (
          <>
            {isLoading && <p className="text-sm text-muted-foreground px-3 py-2">Loading...</p>}
            {!isLoading && error && <p className="text-sm text-muted-foreground px-3 py-2">Could not load conversations</p>}
            {!isLoading && !error && conversations.map(conv => (
              <ConvEntry
                key={conv.id}
                conv={conv}
                active={conv.id === activeId}
                onClick={() => onSelect(conv.id)}
                onDelete={handleDelete}
              />
            ))}
          </>
        )}
      </div>
    </div>
  )
}
