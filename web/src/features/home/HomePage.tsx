import { useForm } from 'react-hook-form'
import { useRandomPrompt, useCreateThread } from '@/hooks/use-thread'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form'
import { useNavigate } from '@tanstack/react-router'

interface FormData {
  message: string
}

export function Home() {
  const navigate = useNavigate()

  const { data: prompt, isLoading: isLoadingPrompt } = useRandomPrompt()
  const createThreadMutation = useCreateThread()

  const form = useForm<FormData>({
    defaultValues: {
      message: '',
    },
  })

  const onSubmit = (data: FormData) => {
    if (prompt) {
      createThreadMutation.mutate(
        {
          initialPrompt: prompt,
          firstUserMessage: data.message,
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
    <div className="flex h-screen flex-col items-center justify-center p-4">
      <div className="w-full max-w-2xl space-y-8">
        <div className="text-center">
          <h1 className="mb-4 text-4xl font-bold">ESL Practice</h1>
          <p className="mb-8 text-xl text-muted-foreground">{prompt}</p>
        </div>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="message"
              rules={{ required: 'Please type a response' }}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Input
                      placeholder="Type your response..."
                      disabled={createThreadMutation.isPending}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button
              type="submit"
              className="w-full"
              disabled={createThreadMutation.isPending}
            >
              {createThreadMutation.isPending ? 'Starting...' : 'Start Conversation'}
            </Button>
          </form>
        </Form>

        {createThreadMutation.isError && (
          <p className="text-center text-destructive">
            Failed to start conversation. Please try again.
          </p>
        )}
      </div>
    </div>
  )
}