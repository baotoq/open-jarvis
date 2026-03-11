'use client'
import { useState } from 'react'
import { Sidebar } from '@/components/Sidebar'
import { ChatArea } from '@/components/ChatArea'
import { useSession } from '@/hooks/useSession'

export default function Home() {
  const { sessionId, persistSessionId, clearSession } = useSession()
  const [activeConvId, setActiveConvId] = useState<string | null>(null)

  // Sync activeConvId from sessionId once resolved from localStorage
  // We use a simple effect-free approach: derive activeConvId from sessionId on first render
  // by letting the ChatArea receive sessionId directly via handleSelect/handleSessionCreated

  const handleSelect = (id: string) => {
    persistSessionId(id)
    setActiveConvId(id)
  }

  const handleNew = () => {
    clearSession()
    setActiveConvId(null)
  }

  const handleSessionCreated = (id: string) => {
    persistSessionId(id)
    setActiveConvId(id)
  }

  // On mount, once sessionId resolves from localStorage, sync activeConvId
  const effectiveConvId = activeConvId ?? sessionId

  return (
    <div className="flex h-screen">
      <Sidebar
        activeId={effectiveConvId}
        onSelect={handleSelect}
        onNew={handleNew}
      />
      <main className="flex-1 flex flex-col overflow-hidden">
        <ChatArea
          sessionId={effectiveConvId}
          onSessionCreated={handleSessionCreated}
        />
      </main>
    </div>
  )
}
