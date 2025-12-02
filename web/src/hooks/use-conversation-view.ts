import { useState, useEffect } from 'react'

export type ConversationViewMode = 'audio' | 'messages'

const VIEW_MODE_KEY = 'conversation-view-mode'

export function useConversationView() {
  const [viewMode, setViewMode] = useState<ConversationViewMode>(() => {
    // Initialize from localStorage
    if (typeof window === 'undefined') return 'audio'
    const stored = localStorage.getItem(VIEW_MODE_KEY)
    return (stored as ConversationViewMode) || 'audio'
  })

  useEffect(() => {
    // Persist to localStorage whenever it changes
    localStorage.setItem(VIEW_MODE_KEY, viewMode)
  }, [viewMode])

  const toggleView = () => {
    setViewMode((prev) => (prev === 'audio' ? 'messages' : 'audio'))
  }

  return { viewMode, setViewMode, toggleView }
}
