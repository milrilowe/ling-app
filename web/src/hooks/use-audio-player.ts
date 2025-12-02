import { useState, useRef, useCallback, useEffect } from 'react'

export type ConversationState = 'idle' | 'recording' | 'ai-thinking' | 'ai-speaking'

export interface AudioPlayerState {
  isPlaying: boolean
  isLoading: boolean
  currentTime: number
  duration: number
  error: string | null
  conversationState: ConversationState
}

export interface AudioPlayerActions {
  play: () => void
  pause: () => void
  seek: (time: number) => void
  setPlaybackRate: (rate: number) => void
  load: (url: string) => void
  setConversationState: (state: ConversationState) => void
}

export function useAudioPlayer(initialUrl?: string) {
  const [state, setState] = useState<AudioPlayerState>({
    isPlaying: false,
    isLoading: false,
    currentTime: 0,
    duration: 0,
    error: null,
    conversationState: 'idle',
  })

  const audioRef = useRef<HTMLAudioElement | null>(null)
  const urlRef = useRef<string | undefined>(initialUrl)

  // Initialize audio element
  useEffect(() => {
    const audio = new Audio()
    audioRef.current = audio

    // Set up event listeners
    audio.addEventListener('loadstart', () => {
      setState(prev => ({ ...prev, isLoading: true, error: null }))
    })

    audio.addEventListener('loadedmetadata', () => {
      setState(prev => ({
        ...prev,
        isLoading: false,
        duration: audio.duration,
      }))
    })

    audio.addEventListener('timeupdate', () => {
      setState(prev => ({ ...prev, currentTime: audio.currentTime }))
    })

    audio.addEventListener('play', () => {
      setState(prev => ({ ...prev, isPlaying: true }))
    })

    audio.addEventListener('pause', () => {
      setState(prev => ({ ...prev, isPlaying: false }))
    })

    audio.addEventListener('ended', () => {
      setState(prev => ({
        ...prev,
        isPlaying: false,
        currentTime: 0,
        conversationState: 'idle' // Reset conversation state when audio ends
      }))
    })

    audio.addEventListener('error', (e) => {
      // Ignore errors when there's no src (initial state)
      if (!audio.src || audio.src === window.location.href) {
        return
      }

      const errorDetails = audio.error ? {
        code: audio.error.code,
        message: audio.error.message,
        mediaErrorCodes: {
          1: 'MEDIA_ERR_ABORTED',
          2: 'MEDIA_ERR_NETWORK',
          3: 'MEDIA_ERR_DECODE',
          4: 'MEDIA_ERR_SRC_NOT_SUPPORTED'
        }
      } : 'Unknown error'
      console.error('Audio playback error:', e, errorDetails, 'URL:', audio.src)
      setState(prev => ({
        ...prev,
        isPlaying: false,
        isLoading: false,
        error: 'Failed to load audio',
      }))
    })

    // Load initial URL if provided
    if (initialUrl) {
      audio.src = initialUrl
      audio.load()
    }

    // Cleanup
    return () => {
      audio.pause()
      audio.src = ''
      audio.load()
    }
  }, [initialUrl])

  const play = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.play().catch(err => {
        console.error('Failed to play audio:', err)
        setState(prev => ({ ...prev, error: 'Failed to play audio' }))
      })
    }
  }, [])

  const pause = useCallback(() => {
    if (audioRef.current) {
      audioRef.current.pause()
    }
  }, [])

  const seek = useCallback((time: number) => {
    if (audioRef.current) {
      audioRef.current.currentTime = time
      setState(prev => ({ ...prev, currentTime: time }))
    }
  }, [])

  const setPlaybackRate = useCallback((rate: number) => {
    if (audioRef.current) {
      audioRef.current.playbackRate = rate
    }
  }, [])

  const load = useCallback((url: string) => {
    if (audioRef.current && url !== urlRef.current) {
      urlRef.current = url
      audioRef.current.src = url
      audioRef.current.load()
    }
  }, [])

  const setConversationState = useCallback((conversationState: ConversationState) => {
    setState(prev => ({ ...prev, conversationState }))
  }, [])

  return {
    ...state,
    play,
    pause,
    seek,
    setPlaybackRate,
    load,
    setConversationState,
  }
}
