import { createFileRoute } from '@tanstack/react-router'
import { PronunciationDashboard } from '@/features/pronunciation/PronunciationDashboard'

export const Route = createFileRoute('/pronunciation')({
  component: PronunciationPage,
})

function PronunciationPage() {
  return (
    <div className="h-screen bg-background overflow-hidden">
      <PronunciationDashboard />
    </div>
  )
}
