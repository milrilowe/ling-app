import { createContext, useContext, type ReactNode } from 'react'
import { useAudioPlayer } from '@/hooks/use-audio-player'
import type { AudioPlayerState, AudioPlayerActions } from '@/hooks/use-audio-player'

type AudioPlayerContextType = AudioPlayerState & AudioPlayerActions

const AudioPlayerContext = createContext<AudioPlayerContextType | undefined>(undefined)

interface AudioPlayerProviderProps {
  children: ReactNode
}

export function AudioPlayerProvider({ children }: AudioPlayerProviderProps) {
  const audioPlayer = useAudioPlayer()

  return (
    <AudioPlayerContext.Provider value={audioPlayer}>
      {children}
    </AudioPlayerContext.Provider>
  )
}

export function useAudioPlayerContext() {
  const context = useContext(AudioPlayerContext)
  if (!context) {
    throw new Error('useAudioPlayerContext must be used within AudioPlayerProvider')
  }
  return context
}
