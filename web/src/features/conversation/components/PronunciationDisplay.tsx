import { useState } from 'react'
import { ChevronDown, Loader2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { PronunciationAnalysis, PhonemeDetail } from '@/lib/api'

interface PronunciationDisplayProps {
  status: 'none' | 'pending' | 'complete' | 'failed'
  analysis?: PronunciationAnalysis
  error?: string
  expectedText?: string
}

type WordStatus = 'correct' | 'partial' | 'wrong' | 'missed'

interface WordWithStatus {
  text: string
  status: WordStatus
}

/**
 * Calculate word-level accuracy from phoneme details.
 * Maps phonemes back to words and determines each word's status.
 */
function getWordStatuses(
  expectedText: string,
  phonemeDetails: PhonemeDetail[]
): WordWithStatus[] {
  const words = expectedText.split(/\s+/).filter(w => w.length > 0)

  if (words.length === 0 || phonemeDetails.length === 0) {
    return words.map(text => ({ text, status: 'correct' as WordStatus }))
  }

  const totalChars = words.reduce((sum, w) => sum + w.replace(/[^\w]/g, '').length, 0)
  const totalPhonemes = phonemeDetails.filter(p => p.type !== 'insert').length

  let phonemeIndex = 0
  const result: WordWithStatus[] = []

  for (const word of words) {
    const wordChars = word.replace(/[^\w]/g, '').length
    const estimatedPhonemes = Math.max(1, Math.round((wordChars / totalChars) * totalPhonemes))
    const wordPhonemes = phonemeDetails.slice(phonemeIndex, phonemeIndex + estimatedPhonemes)
    phonemeIndex += estimatedPhonemes

    const matchCount = wordPhonemes.filter(p => p.type === 'match').length
    const deleteCount = wordPhonemes.filter(p => p.type === 'delete').length

    let status: WordStatus
    if (wordPhonemes.length === 0 || deleteCount === wordPhonemes.length) {
      status = 'missed'
    } else if (matchCount === wordPhonemes.length) {
      status = 'correct'
    } else if (matchCount > 0) {
      status = 'partial'
    } else {
      status = 'wrong'
    }

    result.push({ text: word, status })
  }

  return result
}

/**
 * Get color class based on word status
 */
function getWordColorClass(status: WordStatus): string {
  switch (status) {
    case 'correct':
      return 'text-green-600 dark:text-green-400'
    case 'partial':
      return 'text-yellow-600 dark:text-yellow-400'
    case 'wrong':
      return 'text-yellow-600 dark:text-yellow-400'
    case 'missed':
      return 'text-red-600 dark:text-red-400 line-through'
    default:
      return ''
  }
}

export function PronunciationDisplay({ status, analysis, error, expectedText }: PronunciationDisplayProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  if (status === 'none') {
    return null
  }

  // Pending state
  if (status === 'pending') {
    return (
      <div className="flex items-center gap-2 rounded-b-2xl bg-muted/50 px-4 py-2">
        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        <span className="text-xs text-muted-foreground">Analyzing...</span>
      </div>
    )
  }

  // Failed state
  if (status === 'failed') {
    return (
      <div className="flex items-center gap-2 rounded-b-2xl bg-destructive/10 px-4 py-2">
        <AlertCircle className="h-4 w-4 text-destructive" />
        <span className="text-xs text-destructive">
          {error || 'Analysis failed'}
        </span>
      </div>
    )
  }

  // Complete state
  if (status === 'complete' && analysis) {
    const wordStatuses = expectedText
      ? getWordStatuses(expectedText, analysis.phoneme_details)
      : []

    // Collapsed: just show a small icon on the right
    if (!isExpanded) {
      return (
        <div className="flex justify-end">
          <button
            onClick={() => setIsExpanded(true)}
            className="px-2 py-1 text-muted-foreground hover:text-foreground transition-colors"
            title="Show pronunciation"
          >
            <ChevronDown className="h-3 w-3" />
          </button>
        </div>
      )
    }

    // Expanded: show the colored text analysis (click anywhere to collapse)
    return (
      <button
        onClick={() => setIsExpanded(false)}
        className="rounded-2xl bg-muted/50 mt-1 px-4 py-2 text-left transition-colors hover:bg-muted/70 focus:outline-none w-full"
      >
        <p className="text-sm leading-relaxed break-words">
          {wordStatuses.map((word, idx) => (
            <span key={idx}>
              <span className={getWordColorClass(word.status)}>{word.text}</span>
              {idx < wordStatuses.length - 1 && ' '}
            </span>
          ))}
        </p>
      </button>
    )
  }

  return null
}
