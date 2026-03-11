'use client'
import { useState, useEffect } from 'react'
import { getConversation } from '@/lib/api'

const SESSION_KEY = 'jarvis-session-id'

export function useSession() {
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const stored = localStorage.getItem(SESSION_KEY)
    if (!stored) {
      setIsLoading(false)
      return
    }
    getConversation(stored)
      .then(conv => {
        if (conv) {
          setSessionId(stored)
        } else {
          // Stale ID — clear it; new session will be created on first message
          localStorage.removeItem(SESSION_KEY)
        }
      })
      .catch(() => {
        // Backend unreachable — clear stale ID gracefully
        localStorage.removeItem(SESSION_KEY)
      })
      .finally(() => setIsLoading(false))
  }, [])

  const persistSessionId = (id: string) => {
    localStorage.setItem(SESSION_KEY, id)
    setSessionId(id)
  }

  const clearSession = () => {
    localStorage.removeItem(SESSION_KEY)
    setSessionId(null)
  }

  return { sessionId, isLoading, persistSessionId, clearSession }
}
