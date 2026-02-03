import { useState } from 'react'
import { Loader2, AlertCircle } from 'lucide-react'
import type { PronunciationAnalysis, PhonemeDetail } from '@/lib/api'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

function getFriendlyErrorMessage(error?: string): string {
  if (!error) return 'Analysis failed'
  const code = error.split(':')[0]?.trim()
  switch (code) {
    case 'AUDIO_TOO_SHORT':
      return 'Audio too short. Make sure it\'s at least 1 second long.'
    case 'AUDIO_TOO_LONG':
      return 'Audio too long. Keep it under 30 seconds.'
    case 'AUDIO_DOWNLOAD_FAILED':
    case 'ML_SERVICE_ERROR':
    case 'MODELS_NOT_LOADED':
    case 'PRESIGNED_URL_ERROR':
      return 'Analysis temporarily unavailable. Please try again.'
    default:
      return 'Analysis failed. Please try recording again.'
  }
}

interface PronunciationDisplayProps {
  status: 'none' | 'pending' | 'complete' | 'failed'
  analysis?: PronunciationAnalysis
  error?: string
  expectedText?: string
  isExpanded?: boolean
  onToggleExpand?: () => void
}

type WordStatus = 'correct' | 'partial' | 'wrong' | 'missed'

interface WordWithStatus {
  text: string
  status: WordStatus
  phonemes: PhonemeDetail[]
}

/**
 * Build IPA strings from phoneme details for a word
 */
function buildIpaStrings(phonemes: PhonemeDetail[]): { expected: string; actual: string } {
  const expected = phonemes.map(p => p.expected).filter(Boolean).join('')
  const actual = phonemes.map(p => {
    if (p.type === 'match') return p.expected
    if (p.type === 'substitute') return p.actual
    if (p.type === 'delete') return '' // phoneme was skipped
    if (p.type === 'insert') return p.actual // extra phoneme added
    return p.expected
  }).filter(Boolean).join('')
  return { expected, actual }
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
    return words.map(text => ({ text, status: 'correct' as WordStatus, phonemes: [] }))
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

    result.push({ text: word, status, phonemes: wordPhonemes })
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

export function PronunciationDisplay({
  status,
  analysis,
  error,
  expectedText,
  isExpanded: controlledExpanded,
  onToggleExpand
}: PronunciationDisplayProps) {
  const [internalExpanded, setInternalExpanded] = useState(false)

  // Support both controlled and uncontrolled modes
  const isExpanded = controlledExpanded ?? internalExpanded
  // toggleExpand is available for uncontrolled mode but not used in current UI
  void (onToggleExpand ?? (() => setInternalExpanded(!internalExpanded)))

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
          {getFriendlyErrorMessage(error)}
        </span>
      </div>
    )
  }

  // Complete state
  if (status === 'complete' && analysis) {
    const wordStatuses = expectedText
      ? getWordStatuses(expectedText, analysis.phoneme_details)
      : []

    // Collapsed: return null, chevron is rendered in bubble via renderToggle
    if (!isExpanded) {
      return null
    }

    // Expanded: show the colored text analysis with clickable words
    return (
      <div
        className="rounded-2xl bg-muted/50 mt-1 px-4 py-2 text-left w-fit ml-auto"
      >
        <p className="text-sm leading-relaxed break-words">
          {wordStatuses.map((word, idx) => {
            const ipa = buildIpaStrings(word.phonemes)
            const hasPhonemeData = word.phonemes.length > 0
            const showPopover = hasPhonemeData && word.status !== 'correct'

            if (showPopover) {
              return (
                <span key={idx}>
                  <Popover>
                    <PopoverTrigger asChild>
                      <button
                        type="button"
                        className={cn(
                          getWordColorClass(word.status),
                          'underline decoration-dotted underline-offset-2 cursor-pointer hover:opacity-80'
                        )}
                        onClick={(e) => e.stopPropagation()}
                      >
                        {word.text}
                      </button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto max-w-xs p-3">
                      <div className="space-y-2">
                        <div className="text-xs text-muted-foreground uppercase tracking-wider">
                          {word.text}
                        </div>
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <span className="text-xs text-muted-foreground w-16">Expected:</span>
                            <span className="font-mono text-sm">/{ipa.expected}/</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs text-muted-foreground w-16">You said:</span>
                            <span className={cn('font-mono text-sm', getWordColorClass(word.status))}>
                              /{ipa.actual || '—'}/
                            </span>
                          </div>
                        </div>
                        {word.phonemes.some(p => p.type !== 'match') && (
                          <div className="pt-2 border-t">
                            <div className="text-xs text-muted-foreground mb-1">Phoneme breakdown:</div>
                            <div className="flex flex-wrap gap-1">
                              {word.phonemes.map((p, pIdx) => (
                                <span
                                  key={pIdx}
                                  className={cn(
                                    'font-mono text-xs px-1.5 py-0.5 rounded',
                                    p.type === 'match' && 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
                                    p.type === 'substitute' && 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
                                    p.type === 'delete' && 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 line-through',
                                    p.type === 'insert' && 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                                  )}
                                  title={
                                    p.type === 'match' ? 'Correct' :
                                    p.type === 'substitute' ? `Said /${p.actual}/ instead of /${p.expected}/` :
                                    p.type === 'delete' ? `Skipped /${p.expected}/` :
                                    `Added extra /${p.actual}/`
                                  }
                                >
                                  {p.type === 'substitute' ? `${p.expected}→${p.actual}` :
                                   p.type === 'delete' ? p.expected :
                                   p.type === 'insert' ? `+${p.actual}` :
                                   p.expected}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    </PopoverContent>
                  </Popover>
                  {idx < wordStatuses.length - 1 && ' '}
                </span>
              )
            }

            return (
              <span key={idx}>
                <span className={getWordColorClass(word.status)}>{word.text}</span>
                {idx < wordStatuses.length - 1 && ' '}
              </span>
            )
          })}
        </p>
      </div>
    )
  }

  return null
}
