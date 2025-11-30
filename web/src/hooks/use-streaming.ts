import { useState, useEffect, useCallback } from 'react'

interface UseStreamingOptions {
  speed?: number
  skipOnClick?: boolean
}

export function useStreaming(
  fullText: string,
  options: UseStreamingOptions = {}
) {
  const { speed = 40, skipOnClick = true } = options
  const [displayedText, setDisplayedText] = useState('')
  const [isStreaming, setIsStreaming] = useState(true)

  const skipStreaming = useCallback(() => {
    if (skipOnClick && isStreaming) {
      setDisplayedText(fullText)
      setIsStreaming(false)
    }
  }, [fullText, isStreaming, skipOnClick])

  useEffect(() => {
    if (!fullText) {
      setDisplayedText('')
      setIsStreaming(false)
      return
    }

    setDisplayedText('')
    setIsStreaming(true)

    let currentIndex = 0
    const interval = setInterval(() => {
      if (currentIndex < fullText.length) {
        setDisplayedText(fullText.slice(0, currentIndex + 1))
        currentIndex++
      } else {
        setIsStreaming(false)
        clearInterval(interval)
      }
    }, speed)

    return () => clearInterval(interval)
  }, [fullText, speed])

  return { displayedText, isStreaming, skipStreaming }
}
