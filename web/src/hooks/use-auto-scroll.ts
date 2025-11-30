import { useEffect, useRef } from 'react'

export function useAutoScroll<T extends HTMLElement>(
  dependencies: unknown[] = []
) {
  const ref = useRef<T>(null)
  const shouldAutoScrollRef = useRef(true)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = element
      // Consider "near bottom" as within 100px of the bottom
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 100
      shouldAutoScrollRef.current = isNearBottom
    }

    element.addEventListener('scroll', handleScroll)
    return () => element.removeEventListener('scroll', handleScroll)
  }, [])

  useEffect(() => {
    if (shouldAutoScrollRef.current && ref.current) {
      ref.current.scrollTo({
        top: ref.current.scrollHeight,
        behavior: 'smooth',
      })
    }
  }, dependencies)

  return ref
}
