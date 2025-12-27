import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { PronunciationDashboard } from '@/features/pronunciation/PronunciationDashboard'

export const Route = createFileRoute('/pronunciation')({
  component: PronunciationPage,
})

function PronunciationPage() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="mx-auto max-w-2xl">
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
            <ArrowLeft className="h-4 w-4" />
            Back to app
          </Link>
        </div>

        <div className="mb-8">
          <h1 className="text-2xl font-bold mb-2">Pronunciation Stats</h1>
          <p className="text-muted-foreground">
            Track your pronunciation progress and see which sounds need more practice.
          </p>
        </div>

        <PronunciationDashboard />
      </div>
    </div>
  )
}
