import { useState } from 'react'
import { ChevronDown, ChevronUp, Loader2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { PronunciationAnalysis } from '@/lib/api'

interface PronunciationDisplayProps {
  status: 'none' | 'pending' | 'complete' | 'failed'
  analysis?: PronunciationAnalysis
  error?: string
}

export function PronunciationDisplay({ status, analysis, error }: PronunciationDisplayProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  // Don't render if no analysis needed
  if (status === 'none') {
    return null
  }

  // Calculate accuracy score
  const accuracyScore = analysis
    ? Math.round((analysis.match_count / analysis.phoneme_count) * 100)
    : 0

  // Pending state
  if (status === 'pending') {
    return (
      <div className="mt-2 flex items-center gap-2 rounded-lg border border-border/50 bg-muted/30 px-3 py-2">
        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        <span className="text-xs text-muted-foreground">Analyzing pronunciation...</span>
      </div>
    )
  }

  // Failed state
  if (status === 'failed') {
    return (
      <div className="mt-2 flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2">
        <AlertCircle className="h-4 w-4 text-destructive" />
        <span className="text-xs text-destructive">
          {error || 'Failed to analyze pronunciation'}
        </span>
      </div>
    )
  }

  // Complete state - show expandable analysis
  if (status === 'complete' && analysis) {
    return (
      <div className="mt-2">
        {/* Collapsed header - always visible */}
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className={cn(
            'flex w-full items-center gap-2 rounded-lg border px-3 py-2 text-left transition-colors',
            'border-border/50 bg-muted/30 hover:bg-muted/50',
            isExpanded && 'rounded-b-none border-b-0'
          )}
        >
          {/* Accuracy indicator bar */}
          <div
            className={cn(
              'h-4 w-1 rounded-full',
              accuracyScore >= 80 ? 'bg-green-500' :
              accuracyScore >= 60 ? 'bg-yellow-500' :
              'bg-red-500'
            )}
          />

          {/* Score badge */}
          <span
            className={cn(
              'rounded-full px-2 py-0.5 text-xs font-medium',
              accuracyScore >= 80 ? 'bg-green-500/20 text-green-700 dark:text-green-400' :
              accuracyScore >= 60 ? 'bg-yellow-500/20 text-yellow-700 dark:text-yellow-400' :
              'bg-red-500/20 text-red-700 dark:text-red-400'
            )}
          >
            {accuracyScore}% accuracy
          </span>

          <span className="flex-1" />

          {/* Expand/collapse icon */}
          {isExpanded ? (
            <ChevronUp className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          )}
        </button>

        {/* Expanded details */}
        {isExpanded && (
          <div className="rounded-b-lg border border-t-0 border-border/50 bg-muted/20 px-3 py-3">
            {/* IPA comparison */}
            <div className="mb-3 space-y-1">
              <div className="flex items-baseline gap-2">
                <span className="text-xs text-muted-foreground w-16">Expected:</span>
                <span className="font-mono text-sm">{analysis.expected_ipa}</span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-xs text-muted-foreground w-16">You said:</span>
                <span className="font-mono text-sm">{analysis.audio_ipa}</span>
              </div>
            </div>

            {/* Phoneme breakdown */}
            <div className="mb-2">
              <span className="text-xs text-muted-foreground">Phoneme breakdown:</span>
            </div>
            <div className="flex flex-wrap gap-1">
              {analysis.phoneme_details.map((phoneme, idx) => (
                <PhonemeChip key={idx} phoneme={phoneme} />
              ))}
            </div>

            {/* Stats summary */}
            <div className="mt-3 flex gap-4 text-xs text-muted-foreground">
              <span>{analysis.match_count} correct</span>
              {analysis.substitution_count > 0 && (
                <span className="text-yellow-600 dark:text-yellow-400">
                  {analysis.substitution_count} substituted
                </span>
              )}
              {analysis.deletion_count > 0 && (
                <span className="text-red-600 dark:text-red-400">
                  {analysis.deletion_count} missed
                </span>
              )}
              {analysis.insertion_count > 0 && (
                <span className="text-blue-600 dark:text-blue-400">
                  {analysis.insertion_count} extra
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    )
  }

  return null
}

interface PhonemeChipProps {
  phoneme: {
    expected: string
    actual: string
    type: 'match' | 'substitute' | 'delete' | 'insert'
  }
}

function PhonemeChip({ phoneme }: PhonemeChipProps) {
  const { expected, actual, type } = phoneme

  const baseClasses = 'inline-flex items-center rounded px-1.5 py-0.5 font-mono text-xs'

  switch (type) {
    case 'match':
      return (
        <span className={cn(baseClasses, 'bg-green-500/20 text-green-700 dark:text-green-400')}>
          {expected}
        </span>
      )
    case 'substitute':
      return (
        <span className={cn(baseClasses, 'bg-yellow-500/20 text-yellow-700 dark:text-yellow-400')}>
          <span className="line-through opacity-50">{expected}</span>
          <span className="mx-0.5">→</span>
          <span>{actual}</span>
        </span>
      )
    case 'delete':
      return (
        <span className={cn(baseClasses, 'bg-red-500/20 text-red-700 dark:text-red-400 line-through')}>
          {expected}
        </span>
      )
    case 'insert':
      return (
        <span className={cn(baseClasses, 'bg-blue-500/20 text-blue-700 dark:text-blue-400')}>
          +{actual}
        </span>
      )
    default:
      return null
  }
}
