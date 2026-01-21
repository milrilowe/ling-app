import { createFileRoute } from '@tanstack/react-router'
import { PronunciationDashboard } from '@/features/pronunciation/PronunciationDashboard'

export const Route = createFileRoute('/pronunciation')({
  component: PronunciationPage,
})

function PronunciationPage() {
  return (
    <div className="h-full md:h-auto md:min-h-full bg-background overflow-hidden md:overflow-visible">
      <PronunciationDashboard />
    </div>
  )
}
