import { Card, CardContent } from '@/components/ui/card'

interface PromptCardProps {
  prompt: string
}

export function PromptCard({ prompt }: PromptCardProps) {
  return (
    <Card className="animate-in fade-in slide-in-from-bottom-3 border-2 border-success/20 bg-gradient-to-br from-success/5 to-success/10 duration-500">
      <CardContent className="p-8 text-center">
        <div className="mb-4 text-6xl">💬</div>
        <p className="text-xl font-semibold text-foreground">{prompt}</p>
      </CardContent>
    </Card>
  )
}
