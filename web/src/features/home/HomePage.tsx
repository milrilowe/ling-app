import { useRandomPrompt, useCreateThread } from '@/hooks/use-thread'
import { useNavigate } from '@tanstack/react-router'
import { PromptCard } from './components/PromptCard'
import { AIGreeting } from './components/AIGreeting'
import { ChatInput } from '../chat/components/ChatInput'

export function Home() {
  const navigate = useNavigate()

  const { data: prompt, isLoading: isLoadingPrompt } = useRandomPrompt()
  const createThreadMutation = useCreateThread()

  // AI greeting based on the prompt
  const aiGreeting = prompt
    ? `Hi! ${prompt} I'm here to help you practice. Just start talking!`
    : ''

  const handleStart = (message: string) => {
    if (prompt) {
      createThreadMutation.mutate(
        {
          initialPrompt: aiGreeting,
          firstUserMessage: message,
        },
        {
          onSuccess: (thread) => {
            navigate({ to: '/c/$threadId', params: { threadId: thread.id } })
          },
        }
      )
    }
  }

  if (isLoadingPrompt) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-success/10 via-background to-info/10 p-4">
      <div className="w-full max-w-2xl space-y-8">
        {/* Title */}
        <div className="animate-in fade-in slide-in-from-bottom-1 text-center duration-300">
          <h1 className="mb-2 bg-gradient-to-r from-success to-info bg-clip-text text-5xl font-bold text-transparent">
            Practice English
          </h1>
          <p className="text-lg text-muted-foreground">
            Let's have a conversation!
          </p>
        </div>

        {/* Prompt Card */}
        {prompt && <PromptCard prompt={prompt} />}

        {/* AI Greeting */}
        {aiGreeting && <AIGreeting message={aiGreeting} />}

        {/* Input */}
        <div className="animate-in fade-in slide-in-from-bottom-5 duration-1000">
          <ChatInput
            onSubmit={handleStart}
            disabled={createThreadMutation.isPending}
            placeholder="Type your response to start..."
          />
        </div>

        {/* Error Message */}
        {createThreadMutation.isError && (
          <p className="animate-in fade-in text-center text-destructive">
            Failed to start conversation. Please try again.
          </p>
        )}
      </div>
    </div>
  )
}